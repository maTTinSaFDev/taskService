package transport

import (
	"context"
	"core/model"
	authv1 "database-service/api/auth/v1"
	userv1 "database-service/api/user/v1"
	"database-service/service"
	"errors"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	authv1.UnimplementedAuthPersistenceServiceServer
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (handler *AuthHandler) CreateUser(ctx context.Context, request *authv1.CreateUserRequest) (*userv1.User, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "create user request is nil")
	}
	credential := &model.AuthCredential{
		User:         model.User{Name: strings.TrimSpace(request.Name), Email: strings.TrimSpace(request.Email)},
		PasswordHash: request.PasswordHash,
	}
	user, err := handler.service.CreateUser(ctx, credential)
	if err != nil {
		return nil, authStatusError(err)
	}
	return userToProto(user), nil
}

func (handler *AuthHandler) GetCredentialByEmail(ctx context.Context, request *authv1.GetCredentialByEmailRequest) (*authv1.Credential, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "get credential request is nil")
	}
	credential, err := handler.service.GetCredentialByEmail(ctx, request.Email)
	if err != nil {
		return nil, authStatusError(err)
	}
	return &authv1.Credential{User: userToProto(&credential.User), PasswordHash: credential.PasswordHash}, nil
}

func (handler *AuthHandler) GetUserById(ctx context.Context, request *authv1.GetUserByIdRequest) (*userv1.User, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "get user request is nil")
	}
	user, err := handler.service.GetUserByID(ctx, request.UserId)
	if err != nil {
		return nil, authStatusError(err)
	}
	return userToProto(user), nil
}

func authStatusError(err error) error {
	switch {
	case errors.Is(err, model.ErrEmailAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, model.ErrCredentialNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
