package repository

import (
	"context"
	"core/model"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, credential *model.AuthCredential) (*model.User, error)
	GetCredentialByEmail(ctx context.Context, email string) (*model.AuthCredential, error)
	GetUserByID(ctx context.Context, userID int64) (*model.User, error)
}
