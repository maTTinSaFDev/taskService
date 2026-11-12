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
)

type boardServiceStub struct {
	service.BoardService
	createFunc  func(context.Context, *model.Board) (*model.Board, error)
	getByIDFunc func(context.Context, int64) (*model.Board, error)
}

func (stub *boardServiceStub) Create(ctx context.Context, board *model.Board) (*model.Board, error) {
	return stub.createFunc(ctx, board)
}

func (stub *boardServiceStub) GetById(ctx context.Context, id int64) (*model.Board, error) {
	return stub.getByIDFunc(ctx, id)
}

func TestBoardHandlerCreate(t *testing.T) {
	stub := &boardServiceStub{
		createFunc: func(_ context.Context, board *model.Board) (*model.Board, error) {
			if board.OwnerUserId != 42 || board.Title != "Development" {
				t.Fatalf("board = %+v", board)
			}
			board.Id = 7
			board.DoneColumnId = 11
			return board, nil
		},
	}
	router := mux.NewRouter()
	RegisterBoardRoutes(router, NewBoardHandler(stub))
	request := authenticatedRequest(httptest.NewRequest(http.MethodPost, "/boards", strings.NewReader(`{"owner_user_id":999,"title":"Development"}`)), 42)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"done_column_id":11`) {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestBoardHandlerGetHidesForeignBoard(t *testing.T) {
	stub := &boardServiceStub{
		getByIDFunc: func(context.Context, int64) (*model.Board, error) {
			return &model.Board{Id: 7, OwnerUserId: 99}, nil
		},
	}
	router := mux.NewRouter()
	RegisterBoardRoutes(router, NewBoardHandler(stub))
	request := authenticatedRequest(httptest.NewRequest(http.MethodGet, "/boards/7", nil), 42)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}
