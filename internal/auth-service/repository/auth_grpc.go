package repository

import (
	"context"
	"errors"
	"time"

	"auth-service/model"
	authv1 "database-service/api/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GRPCAuthRepository struct {
	client authv1.AuthPersistenceServiceClient
}

var _ AuthRepository = (*GRPCAuthRepository)(nil)

func NewGRPCAuthRepository(client authv1.AuthPersistenceServiceClient) *GRPCAuthRepository {
	return &GRPCAuthRepository{client: client}
}

func (repository *GRPCAuthRepository) CreateIdentity(ctx context.Context, name string, email string, passwordHash string) (*model.User, error) {
	identity, err := repository.client.CreateIdentity(ctx, &authv1.CreateIdentityRequest{
		Name: name, Email: email, PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, mapAuthGRPCError(err)
	}
	return userFromProto(identity)
}

func (repository *GRPCAuthRepository) GetCredentialByEmail(ctx context.Context, email string) (*model.Credential, error) {
	credential, err := repository.client.GetCredentialByEmail(ctx, &authv1.GetCredentialByEmailRequest{Email: email})
	if err != nil {
		return nil, mapAuthGRPCError(err)
	}
	if credential == nil {
		return nil, errors.New("database-service returned a nil credential")
	}
	user, err := userFromProto(credential.Identity)
	if err != nil {
		return nil, err
	}
	return &model.Credential{User: *user, PasswordHash: credential.PasswordHash}, nil
}

func (repository *GRPCAuthRepository) GetIdentityByID(ctx context.Context, userID int64) (*model.User, error) {
	identity, err := repository.client.GetIdentityById(ctx, &authv1.GetIdentityByIdRequest{UserId: userID})
	if err != nil {
		return nil, mapAuthGRPCError(err)
	}
	return userFromProto(identity)
}

func (repository *GRPCAuthRepository) CreateRefreshSession(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := repository.client.CreateRefreshSession(ctx, &authv1.CreateRefreshSessionRequest{
		UserId: userID, TokenHash: tokenHash, ExpiresAt: timestamppb.New(expiresAt),
	})
	return mapAuthGRPCError(err)
}

func (repository *GRPCAuthRepository) RotateRefreshSession(ctx context.Context, currentTokenHash string, replacementTokenHash string, replacementExpiresAt time.Time) (*model.User, error) {
	identity, err := repository.client.RotateRefreshSession(ctx, &authv1.RotateRefreshSessionRequest{
		CurrentTokenHash:     currentTokenHash,
		ReplacementTokenHash: replacementTokenHash,
		ReplacementExpiresAt: timestamppb.New(replacementExpiresAt),
	})
	if err != nil {
		return nil, mapAuthGRPCError(err)
	}
	return userFromProto(identity)
}

func (repository *GRPCAuthRepository) RevokeRefreshSession(ctx context.Context, tokenHash string) error {
	_, err := repository.client.RevokeRefreshSession(ctx, &authv1.RevokeRefreshSessionRequest{TokenHash: tokenHash})
	return mapAuthGRPCError(err)
}

func userFromProto(identity *authv1.Identity) (*model.User, error) {
	if identity == nil {
		return nil, errors.New("database-service returned a nil identity")
	}
	return &model.User{ID: identity.UserId, Name: identity.Name, Email: identity.Email}, nil
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
	case codes.Unauthenticated:
		return ErrInvalidRefresh
	default:
		return err
	}
}
