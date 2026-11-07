package repository

import (
	"context"
	"core/model"
)

type AuthRepository interface {
	CreateIdentity(ctx context.Context, credential *model.AuthCredential) (*model.AuthIdentity, error)
	GetCredentialByEmail(ctx context.Context, email string) (*model.AuthCredential, error)
	GetIdentityByID(ctx context.Context, userID int64) (*model.AuthIdentity, error)
	CreateRefreshSession(ctx context.Context, session *model.RefreshSession) error
	RotateRefreshSession(ctx context.Context, currentTokenHash string, replacement *model.RefreshSession) (*model.AuthIdentity, error)
	RevokeRefreshSession(ctx context.Context, tokenHash string) error
}
