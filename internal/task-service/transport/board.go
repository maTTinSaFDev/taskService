package transport

import (
	"core/dto"
	"core/model"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"task-service/service"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BoardHandler struct {
	boardService service.BoardService
}

func NewBoardHandler(boardService service.BoardService) *BoardHandler {
	return &BoardHandler{boardService: boardService}
}

func RegisterBoardRoutes(router *mux.Router, handler *BoardHandler) {
	router.HandleFunc("/boards", handler.Create).Methods(http.MethodPost)
	router.HandleFunc("/boards", handler.GetByUserId).Methods(http.MethodGet)
	router.HandleFunc("/boards/{id}", handler.GetById).Methods(http.MethodGet)
	router.HandleFunc("/boards/{board_id}/columns", handler.CreateColumn).Methods(http.MethodPost)
	router.HandleFunc("/boards/{board_id}/columns/{column_id}", handler.UpdateColumn).Methods(http.MethodPut)
	router.HandleFunc("/boards/{board_id}/columns/{column_id}", handler.DeleteColumn).Methods(http.MethodDelete)
}

func (handler *BoardHandler) Create(writer http.ResponseWriter, request *http.Request) {
	userID, ok := requireAuthenticatedUser(writer, request)
	if !ok {
		return
	}
	dtoValue := &dto.BoardCreateDto{}
	if err := json.NewDecoder(request.Body).Decode(dtoValue); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	board, err := handler.boardService.Create(request.Context(), &model.Board{
		OwnerUserId: userID,
		Title:       dtoValue.Title,
	})
	if err != nil {
		writeBoardError(writer, err)
		return
	}
	writeJSON(writer, http.StatusCreated, board)
}

func (handler *BoardHandler) GetById(writer http.ResponseWriter, request *http.Request) {
	userID, ok := requireAuthenticatedUser(writer, request)
	if !ok {
		return
	}
	id, err := positivePathId(request, "id")
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	board, err := handler.boardService.GetById(request.Context(), id)
	if err != nil {
		writeBoardError(writer, err)
		return
	}
	if board.OwnerUserId != userID {
		http.Error(writer, "board not found", http.StatusNotFound)
		return
	}
	writeJSON(writer, http.StatusOK, board)
}

func (handler *BoardHandler) GetByUserId(writer http.ResponseWriter, request *http.Request) {
	userId, ok := requireAuthenticatedUser(writer, request)
	if !ok {
		return
	}
	boards, err := handler.boardService.GetByUserId(request.Context(), userId)
	if err != nil {
		writeBoardError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, boards)
}

func (handler *BoardHandler) CreateColumn(writer http.ResponseWriter, request *http.Request) {
	userID, ok := requireAuthenticatedUser(writer, request)
	if !ok {
		return
	}
	boardId, err := positivePathId(request, "board_id")
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	if !handler.requireOwnedBoard(writer, request, boardId, userID) {
		return
	}
	dtoValue := &dto.BoardColumnCreateDto{}
	if err := json.NewDecoder(request.Body).Decode(dtoValue); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	column, err := handler.boardService.CreateColumn(request.Context(), &model.BoardColumn{
		BoardId:  boardId,
		Title:    dtoValue.Title,
		Position: dtoValue.Position,
	})
	if err != nil {
		writeBoardError(writer, err)
		return
	}
	writeJSON(writer, http.StatusCreated, column)
}

func (handler *BoardHandler) UpdateColumn(writer http.ResponseWriter, request *http.Request) {
	userID, ok := requireAuthenticatedUser(writer, request)
	if !ok {
		return
	}
	boardId, err := positivePathId(request, "board_id")
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	if !handler.requireOwnedBoard(writer, request, boardId, userID) {
		return
	}
	columnId, err := positivePathId(request, "column_id")
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	dtoValue := &dto.BoardColumnUpdateDto{}
	if err := json.NewDecoder(request.Body).Decode(dtoValue); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	column, err := handler.boardService.UpdateColumn(request.Context(), &model.BoardColumn{
		Id:       columnId,
		BoardId:  boardId,
		Title:    dtoValue.Title,
		Position: dtoValue.Position,
	})
	if err != nil {
		writeBoardError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, column)
}

func (handler *BoardHandler) DeleteColumn(writer http.ResponseWriter, request *http.Request) {
	userID, ok := requireAuthenticatedUser(writer, request)
	if !ok {
		return
	}
	boardId, err := positivePathId(request, "board_id")
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	if !handler.requireOwnedBoard(writer, request, boardId, userID) {
		return
	}
	columnId, err := positivePathId(request, "column_id")
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	var targetColumnId int64
	rawTargetColumnId := request.URL.Query().Get("target_column_id")
	if rawTargetColumnId != "" {
		targetColumnId, err = strconv.ParseInt(rawTargetColumnId, 10, 64)
		if err != nil || targetColumnId <= 0 {
			http.Error(writer, "target_column_id must be a positive integer", http.StatusBadRequest)
			return
		}
	}
	if err := handler.boardService.DeleteColumn(request.Context(), boardId, columnId, targetColumnId); err != nil {
		writeBoardError(writer, err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (handler *BoardHandler) requireOwnedBoard(writer http.ResponseWriter, request *http.Request, boardID int64, userID int64) bool {
	board, err := handler.boardService.GetById(request.Context(), boardID)
	if err != nil {
		writeBoardError(writer, err)
		return false
	}
	if board.OwnerUserId != userID {
		http.Error(writer, "board not found", http.StatusNotFound)
		return false
	}
	return true
}

func positivePathId(request *http.Request, name string) (int64, error) {
	value, err := strconv.ParseInt(mux.Vars(request)[name], 10, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return value, nil
}

func writeBoardError(writer http.ResponseWriter, err error) {
	switch status.Code(err) {
	case codes.InvalidArgument, codes.Unknown:
		http.Error(writer, err.Error(), http.StatusBadRequest)
	case codes.NotFound:
		http.Error(writer, err.Error(), http.StatusNotFound)
	case codes.FailedPrecondition:
		http.Error(writer, err.Error(), http.StatusConflict)
	case codes.Unavailable:
		http.Error(writer, "database service unavailable", http.StatusBadGateway)
	default:
		http.Error(writer, "internal server error", http.StatusInternalServerError)
	}
}
