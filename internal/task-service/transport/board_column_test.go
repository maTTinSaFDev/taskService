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

type boardColumnServiceStub struct {
	service.BoardColumnService
	reorderFunc func(context.Context, int64, []int64) error
	deleteFunc  func(context.Context, int64, int64, int64) error
}

func (stub *boardColumnServiceStub) Reorder(ctx context.Context, boardId int64, columnIds []int64) error {
	return stub.reorderFunc(ctx, boardId, columnIds)
}

func (stub *boardColumnServiceStub) Delete(ctx context.Context, boardId int64, id int64, targetColumnId int64) error {
	return stub.deleteFunc(ctx, boardId, id, targetColumnId)
}

func TestBoardColumnHandlerReorderPassesCompleteOrder(t *testing.T) {
	boardStub := &boardServiceStub{
		getByIDFunc: func(context.Context, int64) (*model.Board, error) {
			return &model.Board{Id: 7, OwnerUserId: 42}, nil
		},
	}
	columnStub := &boardColumnServiceStub{
		reorderFunc: func(_ context.Context, boardId int64, columnIds []int64) error {
			if boardId != 7 {
				t.Fatalf("board id = %d, want 7", boardId)
			}
			want := []int64{11, 8, 9}
			if len(columnIds) != len(want) {
				t.Fatalf("column ids = %v, want %v", columnIds, want)
			}
			for index := range want {
				if columnIds[index] != want[index] {
					t.Fatalf("column ids = %v, want %v", columnIds, want)
				}
			}
			return nil
		},
	}
	router := mux.NewRouter()
	RegisterBoardColumnRoutes(router, NewBoardColumnHandler(boardStub, columnStub))
	request := authenticatedRequest(httptest.NewRequest(http.MethodPut, "/boards/7/columns/order", strings.NewReader(`{"column_ids":[11,8,9]}`)), 42)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestBoardColumnHandlerDeleteDoneColumnReturnsConflict(t *testing.T) {
	boardStub := &boardServiceStub{
		getByIDFunc: func(context.Context, int64) (*model.Board, error) {
			return &model.Board{Id: 7, OwnerUserId: 42}, nil
		},
	}
	columnStub := &boardColumnServiceStub{
		deleteFunc: func(context.Context, int64, int64, int64) error {
			return status.Error(codes.FailedPrecondition, "done column cannot be deleted")
		},
	}
	router := mux.NewRouter()
	RegisterBoardColumnRoutes(router, NewBoardColumnHandler(boardStub, columnStub))
	request := authenticatedRequest(httptest.NewRequest(http.MethodDelete, "/boards/7/columns/11", nil), 42)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
