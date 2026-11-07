package service

import (
	"context"
	"core/model"
	"errors"
	"strings"
	"time"

	"database-service/repository"
)

type AuthService interface {
	CreateIdentity(ctx context.Context, credential *model.AuthCredential) (*model.AuthIdentity, error)
	GetCredentialByEmail(ctx context.Context, email string) (*model.AuthCredential, error)
	GetIdentityByID(ctx context.Context, userID int64) (*model.AuthIdentity, error)
	CreateRefreshSession(ctx context.Context, session *model.RefreshSession) error
	RotateRefreshSession(ctx context.Context, currentTokenHash string, replacement *model.RefreshSession) (*model.AuthIdentity, error)
	RevokeRefreshSession(ctx context.Context, tokenHash string) error
}

type AuthServiceImpl struct {
	repository repository.AuthRepository
}

func NewAuthService(repository repository.AuthRepository) *AuthServiceImpl {
	return &AuthServiceImpl{repository: repository}
}

func (service *AuthServiceImpl) CreateIdentity(ctx context.Context, credential *model.AuthCredential) (*model.AuthIdentity, error) {
	if credential == nil || strings.TrimSpace(credential.Identity.Name) == "" || strings.TrimSpace(credential.Identity.Email) == "" || credential.PasswordHash == "" {
		return nil, errors.New("name, email and password hash are required")
	}
	return service.repository.CreateIdentity(ctx, credential)
}

func (service *AuthServiceImpl) GetCredentialByEmail(ctx context.Context, email string) (*model.AuthCredential, error) {
	if strings.TrimSpace(email) == "" {
		return nil, errors.New("email is required")
	}
	return service.repository.GetCredentialByEmail(ctx, email)
}

func (service *AuthServiceImpl) GetIdentityByID(ctx context.Context, userID int64) (*model.AuthIdentity, error) {
	if userID <= 0 {
		return nil, errors.New("user id must be greater than zero")
	}
	return service.repository.GetIdentityByID(ctx, userID)
}

func (service *AuthServiceImpl) CreateRefreshSession(ctx context.Context, session *model.RefreshSession) error {
	if err := validateRefreshSession(session, true); err != nil {
		return err
	}
	return service.repository.CreateRefreshSession(ctx, session)
}

func (service *AuthServiceImpl) RotateRefreshSession(ctx context.Context, currentTokenHash string, replacement *model.RefreshSession) (*model.AuthIdentity, error) {
	if strings.TrimSpace(currentTokenHash) == "" {
		return nil, errors.New("current token hash is required")
	}
	if err := validateRefreshSession(replacement, false); err != nil {
		return nil, err
	}
	return service.repository.RotateRefreshSession(ctx, currentTokenHash, replacement)
}

func (service *AuthServiceImpl) RevokeRefreshSession(ctx context.Context, tokenHash string) error {
	if strings.TrimSpace(tokenHash) == "" {
		return errors.New("token hash is required")
	}
	return service.repository.RevokeRefreshSession(ctx, tokenHash)
}

func validateRefreshSession(session *model.RefreshSession, requireUser bool) error {
	if session == nil || (requireUser && session.UserID <= 0) || strings.TrimSpace(session.TokenHash) == "" || session.ExpiresAt.Before(time.Now()) {
		return errors.New("valid refresh session is required")
	}
	return nil
}
