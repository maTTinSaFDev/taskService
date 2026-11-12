package repository

import (
	"context"
	"errors"

	"auth-service/model"
	authv1 "database-service/api/auth/v1"
	userv1 "database-service/api/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCAuthRepository struct {
	client authv1.AuthPersistenceServiceClient
}

func NewGRPCAuthRepository(client authv1.AuthPersistenceServiceClient) *GRPCAuthRepository {
	return &GRPCAuthRepository{client: client}
}

func (repository *GRPCAuthRepository) CreateUser(ctx context.Context, name string, email string, passwordHash string) (*model.User, error) {
	user, err := repository.client.CreateUser(ctx, &authv1.CreateUserRequest{
		Name: name, Email: email, PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, mapAuthGRPCError(err)
	}
	return userFromProto(user)
}

func (repository *GRPCAuthRepository) GetCredentialByEmail(ctx context.Context, email string) (*model.Credential, error) {
	credential, err := repository.client.GetCredentialByEmail(ctx, &authv1.GetCredentialByEmailRequest{Email: email})
	if err != nil {
		return nil, mapAuthGRPCError(err)
	}
	if credential == nil {
		return nil, errors.New("database-service returned a nil credential")
	}
	user, err := userFromProto(credential.User)
	if err != nil {
		return nil, err
	}
	return &model.Credential{User: *user, PasswordHash: credential.PasswordHash}, nil
}

func (repository *GRPCAuthRepository) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
	user, err := repository.client.GetUserById(ctx, &authv1.GetUserByIdRequest{UserId: userID})
	if err != nil {
		return nil, mapAuthGRPCError(err)
	}
	return userFromProto(user)
}

func userFromProto(user *userv1.User) (*model.User, error) {
	if user == nil {
		return nil, errors.New("database-service returned a nil user")
	}
	return &model.User{ID: user.Id, Name: user.Name, Email: user.Email}, nil
}

func mapAuthGRPCError(err error) error {
	if err == nil {
		return nil
	}
	switch status.Code(err) {
	case codes.AlreadyExists:
		return ErrEmailExists
	case codes.NotFound:
		return ErrNotFound
	default:
		return err
	}
}
