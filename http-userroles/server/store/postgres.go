package store

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudmachinery/apps/http-userroles/contracts"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var _ Store = (*PostgresStore)(nil)

type PostgresStore struct {
	db *pgx.Conn
}

type queryable interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...interface{}) pgx.Row
}

func NewPostgresStore(ctx context.Context, connectionString string) (*PostgresStore, error) {
	db, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS users (
            email varchar(32) PRIMARY KEY,
            full_name varchar(32)
        );
        CREATE TABLE IF NOT EXISTS roles (
            email varchar(32) REFERENCES users(email) ON DELETE CASCADE, 
            role varchar(12) NOT NULL,
            PRIMARY KEY(email, role)
        );
        CREATE INDEX IF NOT EXISTS idx_roles_role ON roles(role);
        CREATE INDEX IF NOT EXISTS idx_roles_email ON roles(email);`)
	if err != nil {
		return nil, err
	}
	return &PostgresStore{
		db: db,
	}, nil
}

func (pg *PostgresStore) Close(ctx context.Context) error {
	return pg.db.Close(ctx)
}

func (pg *PostgresStore) GetUsers(ctx context.Context) ([]*contracts.User, error) {
	rows, err := pg.db.Query(ctx, `SELECT u.email, u.full_name, r.role 
		FROM users u
		LEFT JOIN roles r ON u.email = r.email;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users, err := scanUsers(rows)
	if err != nil {
		return nil, err
	}

	return sortUsers(users), nil
}

func (pg *PostgresStore) GetUser(ctx context.Context, email string) (*contracts.User, error) {
	return pg.getUser(ctx, pg.db, email)
}

func (pg *PostgresStore) getUser(ctx context.Context, q queryable, email string) (*contracts.User, error) {
	rows, err := q.Query(ctx, `SELECT u.email, u.full_name, r.role 
		FROM users u
		LEFT JOIN roles r ON u.email = r.email
		WHERE u.email = $1;`, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user, err := scanUser(rows)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (pg *PostgresStore) GetUsersByRole(ctx context.Context, role string) ([]*contracts.User, error) {
	rows, err := pg.db.Query(ctx, `SELECT u.email, u.full_name, r.role 
		FROM users u
		LEFT JOIN roles r ON u.email = r.email
		WHERE u.email IN (SELECT DISTINCT(email) FROM roles WHERE role = $1);`, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users, err := scanUsers(rows)
	if err != nil {
		return nil, err
	}

	return sortUsers(users), nil
}

func (pg *PostgresStore) CreateUser(ctx context.Context, user *contracts.User) error {
	err := pg.inTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "INSERT INTO users VALUES ($1, $2);", user.Email, user.FullName)
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
				return ErrUserAlreadyExists
			}
			return fmt.Errorf("insert user: %w", err)
		}

		prevRoles := NewSet[string]()
		newRoles := NewSet(user.Roles...)
		return adjustRoles(user.Email, prevRoles, newRoles, pg.deleteRoles(ctx, tx), pg.addRoles(ctx, tx))
	})
	return err
}

func (pg *PostgresStore) UpdateUser(ctx context.Context, user *contracts.User) error {
	err := pg.inTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		sqlUser, err := pg.getUser(ctx, tx, user.Email)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, "UPDATE users SET full_name = $1 WHERE email = $2;", user.FullName, user.Email)
		if err != nil {
			return fmt.Errorf("update user: %w", err)
		}

		prevRoles := NewSet(sqlUser.Roles...)
		newRoles := NewSet(user.Roles...)
		return adjustRoles(user.Email, prevRoles, newRoles, pg.deleteRoles(ctx, tx), pg.addRoles(ctx, tx))
	})
	return err
}

func (pg *PostgresStore) DeleteUser(ctx context.Context, email string) error {
	r, err := pg.db.Exec(ctx, "DELETE FROM users WHERE email = $1;", email)
	if err != nil {
		return fmt.Errorf("delete from users: %w", err)
	}

	return errIfNoRowsAffected(r, ErrUserNotFound)
}

func scanUsers(rows pgx.Rows) ([]*contracts.User, error) {
	defer rows.Close()

	userMap := map[string]*contracts.User{}
	for rows.Next() {
		var (
			email, fullName string
			role            *string
		)

		if err := rows.Scan(&email, &fullName, &role); err != nil {
			return nil, err
		}

		user, ok := userMap[email]
		if !ok {
			user = &contracts.User{Email: email, FullName: fullName}
			userMap[email] = user
		}

		if role != nil {
			user.Roles = append(user.Roles, *role)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	users := make([]*contracts.User, 0, len(userMap))
	for _, user := range userMap {
		users = append(users, user)
	}

	return users, nil
}

func scanUser(rows pgx.Rows) (*contracts.User, error) {
	users, err := scanUsers(rows)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, ErrUserNotFound
	}

	return users[0], nil
}

func (pg *PostgresStore) addRoles(ctx context.Context, tx pgx.Tx) ChangeRoleFunc {
	return func(email string, rolesToAdd Set[string]) error {
		for _, role := range rolesToAdd.Elements() {
			_, err := tx.Exec(ctx, "INSERT INTO roles VALUES ($1, $2);", email, role)
			if err != nil {
				return fmt.Errorf("add role %q: %w", role, err)
			}
		}

		return nil
	}
}

func (pg *PostgresStore) deleteRoles(ctx context.Context, tx pgx.Tx) ChangeRoleFunc {
	return func(email string, rolesToRemove Set[string]) error {
		for _, role := range rolesToRemove.Elements() {
			_, err := tx.Exec(ctx, "DELETE FROM roles WHERE email = $1 AND role = $2", email, role)
			if err != nil {
				return fmt.Errorf("delete user: %w", err)
			}
		}

		return nil
	}
}

func (pg *PostgresStore) inTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := pg.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if txErr := tx.Rollback(ctx); txErr != nil {
				log.Printf("rollback transaction failed: %s", txErr)
			}
		} else {
			err = tx.Commit(ctx)
		}
	}()

	return fn(ctx, tx)
}

func errIfNoRowsAffected(ct pgconn.CommandTag, err error) error {
	if ct.RowsAffected() == 0 {
		return err
	}

	return nil
}
