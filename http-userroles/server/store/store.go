package store

import (
	"context"
	"errors"

	"github.com/cloudmachinery/apps/http-userroles/contracts"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

type Store interface {
	Close(ctx context.Context) error
	GetUsers(ctx context.Context) ([]*contracts.User, error)
	GetUser(ctx context.Context, email string) (*contracts.User, error)
	GetUsersByRole(ctx context.Context, role string) ([]*contracts.User, error)
	CreateUser(ctx context.Context, user *contracts.User) error
	UpdateUser(ctx context.Context, user *contracts.User) error
	DeleteUser(ctx context.Context, email string) error
}
