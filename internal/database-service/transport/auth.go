package transport

import (
	"context"
	"core/model"
	authv1 "database-service/api/auth/v1"
	"database-service/service"
	"errors"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthHandler struct {
	authv1.UnimplementedAuthPersistenceServiceServer
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (handler *AuthHandler) CreateIdentity(ctx context.Context, request *authv1.CreateIdentityRequest) (*authv1.Identity, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "create identity request is nil")
	}
	credential := &model.AuthCredential{
		Identity:     model.AuthIdentity{Name: strings.TrimSpace(request.Name), Email: strings.TrimSpace(request.Email)},
		PasswordHash: request.PasswordHash,
	}
	identity, err := handler.service.CreateIdentity(ctx, credential)
	if err != nil {
		return nil, authStatusError(err)
	}
	return identityToProto(identity), nil
}

func (handler *AuthHandler) GetCredentialByEmail(ctx context.Context, request *authv1.GetCredentialByEmailRequest) (*authv1.Credential, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "get credential request is nil")
	}
	credential, err := handler.service.GetCredentialByEmail(ctx, request.Email)
	if err != nil {
		return nil, authStatusError(err)
	}
	return &authv1.Credential{Identity: identityToProto(&credential.Identity), PasswordHash: credential.PasswordHash}, nil
}

func (handler *AuthHandler) GetIdentityById(ctx context.Context, request *authv1.GetIdentityByIdRequest) (*authv1.Identity, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "get identity request is nil")
	}
	identity, err := handler.service.GetIdentityByID(ctx, request.UserId)
	if err != nil {
		return nil, authStatusError(err)
	}
	return identityToProto(identity), nil
}

func (handler *AuthHandler) CreateRefreshSession(ctx context.Context, request *authv1.CreateRefreshSessionRequest) (*emptypb.Empty, error) {
	if request == nil || request.ExpiresAt == nil {
		return nil, status.Error(codes.InvalidArgument, "refresh session and expiry are required")
	}
	if err := request.ExpiresAt.CheckValid(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "refresh session expiry is invalid")
	}
	err := handler.service.CreateRefreshSession(ctx, &model.RefreshSession{
		UserID: request.UserId, TokenHash: request.TokenHash, ExpiresAt: request.ExpiresAt.AsTime(),
	})
	if err != nil {
		return nil, authStatusError(err)
	}
	return &emptypb.Empty{}, nil
}

func (handler *AuthHandler) RotateRefreshSession(ctx context.Context, request *authv1.RotateRefreshSessionRequest) (*authv1.Identity, error) {
	if request == nil || request.ReplacementExpiresAt == nil {
		return nil, status.Error(codes.InvalidArgument, "refresh rotation request and expiry are required")
	}
	if err := request.ReplacementExpiresAt.CheckValid(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "replacement expiry is invalid")
	}
	identity, err := handler.service.RotateRefreshSession(ctx, request.CurrentTokenHash, &model.RefreshSession{
		TokenHash: request.ReplacementTokenHash, ExpiresAt: request.ReplacementExpiresAt.AsTime(),
	})
	if err != nil {
		return nil, authStatusError(err)
	}
	return identityToProto(identity), nil
}

func (handler *AuthHandler) RevokeRefreshSession(ctx context.Context, request *authv1.RevokeRefreshSessionRequest) (*emptypb.Empty, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "revoke refresh session request is nil")
	}
	if err := handler.service.RevokeRefreshSession(ctx, request.TokenHash); err != nil {
		return nil, authStatusError(err)
	}
	return &emptypb.Empty{}, nil
}

func identityToProto(identity *model.AuthIdentity) *authv1.Identity {
	return &authv1.Identity{UserId: identity.UserID, Name: identity.Name, Email: identity.Email}
}

func authStatusError(err error) error {
	switch {
	case errors.Is(err, model.ErrEmailAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, model.ErrCredentialNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, model.ErrRefreshSessionNotFound), errors.Is(err, model.ErrRefreshSessionInvalid):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
