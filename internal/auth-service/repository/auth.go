package repository

import (
	"context"
	"errors"
	"time"

	"auth-service/model"
)

var (
	ErrEmailExists    = errors.New("email already exists")
	ErrNotFound       = errors.New("identity not found")
	ErrInvalidRefresh = errors.New("refresh session is invalid")
)

type AuthRepository interface {
	CreateIdentity(ctx context.Context, name string, email string, passwordHash string) (*model.User, error)
	GetCredentialByEmail(ctx context.Context, email string) (*model.Credential, error)
	GetIdentityByID(ctx context.Context, userID int64) (*model.User, error)
	CreateRefreshSession(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	RotateRefreshSession(ctx context.Context, currentTokenHash string, replacementTokenHash string, replacementExpiresAt time.Time) (*model.User, error)
	RevokeRefreshSession(ctx context.Context, tokenHash string) error
}
