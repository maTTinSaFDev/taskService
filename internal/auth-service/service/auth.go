package service

import (
	"context"
	"errors"

	"auth-service/model"
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrEmailExists        = errors.New("email is already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidRefresh     = errors.New("refresh session is invalid")
	ErrUnauthorized       = errors.New("unauthorized")
)

type AuthService interface {
	Register(ctx context.Context, name string, email string, password string) (*model.Session, error)
	Login(ctx context.Context, email string, password string) (*model.Session, error)
	Refresh(ctx context.Context, currentRefreshToken string) (*model.Session, error)
	Logout(ctx context.Context, refreshToken string) error
	Me(ctx context.Context, accessToken string) (*model.User, error)
}
