package repository_test

import (
	"context"
	"core/model"
	userv1 "database-service/api/user/v1"
	"task-service/repository"
	"testing"

	"google.golang.org/grpc"
)

type userServiceClientStub struct {
	userv1.UserServiceClient
	createUserFunc func(context.Context, *userv1.CreateUserRequest, ...grpc.CallOption) (*userv1.CreateUserResponse, error)
}

func (stub *userServiceClientStub) CreateUser(ctx context.Context, request *userv1.CreateUserRequest, options ...grpc.CallOption) (*userv1.CreateUserResponse, error) {
	return stub.createUserFunc(ctx, request, options...)
}

func TestGRPCUserRepositoryCreateMapsRequestAndResponse(t *testing.T) {
	ctx := context.Background()
	client := &userServiceClientStub{
		createUserFunc: func(gotContext context.Context, request *userv1.CreateUserRequest, _ ...grpc.CallOption) (*userv1.CreateUserResponse, error) {
			if gotContext != ctx {
				t.Fatal("CreateUser did not preserve context")
			}
			if request.Name != "Vadim" || request.Email != "vadim@example.com" {
				t.Fatalf("request = %+v", request)
			}
			return &userv1.CreateUserResponse{User: &userv1.User{Id: 7, Name: "Vadim", Email: "vadim@example.com"}}, nil
		},
	}

	user, err := repository.NewGRPCUserRepository(client).Create(ctx, &model.User{Name: "Vadim", Email: "vadim@example.com"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if user.Id != 7 || user.Name != "Vadim" || user.Email != "vadim@example.com" {
		t.Fatalf("user = %+v", user)
	}
}
