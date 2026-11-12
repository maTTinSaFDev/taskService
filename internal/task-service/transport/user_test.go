package transport

import (
	"context"
	"core/model"
	"net/http"
	"net/http/httptest"
	"strings"
	"task-service/service"
	"testing"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type httpUserServiceStub struct {
	service.UserService
	createFunc func(context.Context, *model.User) (*model.User, error)
	deleteFunc func(context.Context, int64) error
}

func (stub *httpUserServiceStub) Create(ctx context.Context, user *model.User) (*model.User, error) {
	return stub.createFunc(ctx, user)
}

func (stub *httpUserServiceStub) Delete(ctx context.Context, id int64) error {
	return stub.deleteFunc(ctx, id)
}

func TestUserHandlerCreateRouteIsNotExposed(t *testing.T) {
	stub := &httpUserServiceStub{}
	router := mux.NewRouter()
	RegisterUserRoutes(router, NewUserHandler(stub))
	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Vadim"}`))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestUserHandlerDeleteReferencedUserReturnsConflict(t *testing.T) {
	stub := &httpUserServiceStub{
		deleteFunc: func(context.Context, int64) error {
			return status.Error(codes.FailedPrecondition, "user is referenced")
		},
	}
	router := mux.NewRouter()
	RegisterUserRoutes(router, NewUserHandler(stub))
	request := authenticatedRequest(httptest.NewRequest(http.MethodDelete, "/users/7", nil), 7)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
