package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"github.com/cloudmachinery/apps/http-userroles/contracts"
)

var _ Store = (*BoltDBStore)(nil)

var (
	usersBucketName = []byte("users")
	rolesBucketName = []byte("roles")
)

type BoltDBStore struct {
	db *bolt.DB
}

func NewBoltStore(filePath string) (*BoltDBStore, error) {
	db, err := bolt.Open(filePath, 0o600, nil)
	if err != nil {
		return nil, err
	}

	if err = initBuckets(db); err != nil {
		return nil, err
	}

	return &BoltDBStore{
		db: db,
	}, nil
}

func initBuckets(db *bolt.DB) error {
	err := db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(usersBucketName); err != nil {
			return fmt.Errorf("unable to create bucket %q: %w", usersBucketName, err)
		}
		if _, err := tx.CreateBucketIfNotExists(rolesBucketName); err != nil {
			return fmt.Errorf("unable to create bucket %q: %w", rolesBucketName, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error initializing buckets: %w", err)
	}
	return nil
}

func (b *BoltDBStore) Close(_ context.Context) error {
	return b.db.Close()
}

func (b *BoltDBStore) GetUsers(_ context.Context) ([]*contracts.User, error) {
	var users []*contracts.User

	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := usersBucket(tx)

		return bucket.ForEach(func(k, v []byte) error {
			var user *contracts.User

			if err := json.Unmarshal(v, &user); err != nil {
				return fmt.Errorf("error unmarshalling user %q: %w", user.Email, err)
			}
			users = append(users, user)
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get users: %w", err)
	}

	return sortUsers(users), nil
}

func (b *BoltDBStore) GetUser(_ context.Context, email string) (*contracts.User, error) {
	var user *contracts.User

	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := usersBucket(tx)
		v := bucket.Get([]byte(email))
		if v == nil {
			return ErrUserNotFound
		}
		return json.Unmarshal(v, &user)
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (b *BoltDBStore) GetUsersByRole(_ context.Context, role string) ([]*contracts.User, error) {
	var users []*contracts.User

	err := b.db.View(func(tx *bolt.Tx) error {
		roleBucket := roleNameBucket(tx, role)
		if roleBucket == nil {
			return nil
		}

		usersBucket := usersBucket(tx)
		return roleBucket.ForEach(func(k, v []byte) error {
			userBytes := usersBucket.Get(k)
			if userBytes == nil {
				log.Printf("user %q not found", k)
				return nil
			}

			var user *contracts.User
			if err := json.Unmarshal(userBytes, &user); err != nil {
				return fmt.Errorf("error unmarshalling user %q: %w", k, err)
			}

			users = append(users, user)
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("error getting users by role %q: %w", role, err)
	}

	return sortUsers(users), nil
}

func (b *BoltDBStore) CreateUser(_ context.Context, user *contracts.User) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := usersBucket(tx)

		if bucket.Get([]byte(user.Email)) != nil {
			return ErrUserAlreadyExists
		}

		userBytes, err := json.Marshal(user)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(user.Email), userBytes)
		if err != nil {
			return err
		}

		rolesSet := NewSet(user.Roles...)
		err = adjustRoles(user.Email, NewSet[string](), rolesSet, b.deleteRoles(tx), b.addRoles(tx))
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (b *BoltDBStore) UpdateUser(_ context.Context, user *contracts.User) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := usersBucket(tx)

		prevUserBytes := bucket.Get([]byte(user.Email))
		if prevUserBytes == nil {
			return ErrUserNotFound
		}

		var prevUser *contracts.User
		if err := json.Unmarshal(prevUserBytes, &prevUser); err != nil {
			return err
		}

		userBytes, err := json.Marshal(&user)
		if err != nil {
			return err
		}
		if err := bucket.Put([]byte(user.Email), userBytes); err != nil {
			return err
		}

		prevRoles := NewSet(prevUser.Roles...)
		newRoles := NewSet(user.Roles...)
		if err := adjustRoles(user.Email, prevRoles, newRoles, b.deleteRoles(tx), b.addRoles(tx)); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (b *BoltDBStore) DeleteUser(_ context.Context, email string) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := usersBucket(tx)

		userBytes := bucket.Get([]byte(email))
		if userBytes == nil {
			return ErrUserNotFound
		}

		var user *contracts.User
		if err := json.Unmarshal(userBytes, &user); err != nil {
			return fmt.Errorf("error unmarshalling user %q: %w", email, err)
		}

		if err := bucket.Delete([]byte(user.Email)); err != nil {
			return fmt.Errorf("unable to delete user %q from users bucket %q: %w", user.Email, usersBucketName, err)
		}

		prevRoles := NewSet(user.Roles...)
		newRoles := NewSet[string]()
		if err := adjustRoles(user.Email, prevRoles, newRoles, b.deleteRoles(tx), b.addRoles(tx)); err != nil {
			return fmt.Errorf("error adjusting roles %v for user %q: %w", user.Roles, user.Email, err)
		}

		return nil
	})

	return err
}

func (b *BoltDBStore) addRoles(tx *bolt.Tx) ChangeRoleFunc {
	return func(email string, rolesToAdd Set[string]) error {
		for _, role := range rolesToAdd.Elements() {
			roleNameBucket, err := rolesBucket(tx).CreateBucketIfNotExists([]byte(role))
			if err != nil {
				return fmt.Errorf("error creating bucket for role %q: %w", role, err)
			}
			if err := roleNameBucket.Put([]byte(email), []byte("")); err != nil {
				return fmt.Errorf("unable to put user %q into bucket %q: %w", email, role, err)
			}
		}
		return nil
	}
}

func (b *BoltDBStore) deleteRoles(tx *bolt.Tx) ChangeRoleFunc {
	return func(email string, rolesToRemove Set[string]) error {
		for _, role := range rolesToRemove.Elements() {
			roleNameBucket := roleNameBucket(tx, role)

			if roleNameBucket == nil {
				fmt.Printf("user %q has role %q that does not exist", email, role)
				continue
			}

			user := roleNameBucket.Get([]byte(email))
			if roleNameBucket.Stats().KeyN == 1 && user != nil {
				if err := rolesBucket(tx).DeleteBucket([]byte(role)); err != nil {
					return fmt.Errorf("unable to delete bucket %q: %w", role, err)
				}
				continue
			}

			if err := roleNameBucket.Delete([]byte(email)); err != nil {
				return fmt.Errorf("unable to delete role %q for user %q: %w", role, email, err)
			}
		}

		return nil
	}
}

func usersBucket(tx *bolt.Tx) *bolt.Bucket {
	return tx.Bucket(usersBucketName)
}

func rolesBucket(tx *bolt.Tx) *bolt.Bucket {
	return tx.Bucket(rolesBucketName)
}

func roleNameBucket(tx *bolt.Tx, roleName string) *bolt.Bucket {
	return rolesBucket(tx).Bucket([]byte(roleName))
}
