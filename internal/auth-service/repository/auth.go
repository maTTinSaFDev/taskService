package repository

import (
	"context"
	"errors"
	"time"

	"auth-service/model"
)

var (
	ErrEmailExists    = errors.New("email already exists")
	ErrNotFound       = errors.New("user not found")
	ErrInvalidRefresh = errors.New("refresh session is invalid")
)

type AuthRepository interface {
	CreateUser(ctx context.Context, name string, email string, passwordHash string) (*model.User, error)
	GetCredentialByEmail(ctx context.Context, email string) (*model.Credential, error)
	GetUserByID(ctx context.Context, userID int64) (*model.User, error)
}

type SessionRepository interface {
	Create(ctx context.Context, user *model.User, tokenHash string, ttl time.Duration) error
	Rotate(ctx context.Context, currentTokenHash string, replacementTokenHash string, ttl time.Duration) (*model.User, error)
	Revoke(ctx context.Context, tokenHash string) error
}
