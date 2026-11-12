package service

import (
	"context"
	"core/model"
	"database-service/repository"
	"errors"
	"strings"
)

type AuthServiceImpl struct {
	repository repository.AuthRepository
}

func NewAuthService(repository repository.AuthRepository) *AuthServiceImpl {
	return &AuthServiceImpl{repository: repository}
}

func (service *AuthServiceImpl) CreateUser(ctx context.Context, credential *model.AuthCredential) (*model.User, error) {
	if credential == nil || strings.TrimSpace(credential.User.Name) == "" || strings.TrimSpace(credential.User.Email) == "" || credential.PasswordHash == "" {
		return nil, errors.New("name, email and password hash are required")
	}
	return service.repository.CreateUser(ctx, credential)
}

func (service *AuthServiceImpl) GetCredentialByEmail(ctx context.Context, email string) (*model.AuthCredential, error) {
	if strings.TrimSpace(email) == "" {
		return nil, errors.New("email is required")
	}
	return service.repository.GetCredentialByEmail(ctx, email)
}

func (service *AuthServiceImpl) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
	if userID <= 0 {
		return nil, errors.New("user id must be greater than zero")
	}
	return service.repository.GetUserByID(ctx, userID)
}
