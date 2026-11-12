package transport

import (
	"core/dto"
	"core/model"
	"encoding/json"
	"net/http"
	"strconv"
	"task-service/service"

	"github.com/gorilla/mux"
)

type BoardColumnHandler struct {
	boardService       service.BoardService
	boardColumnService service.BoardColumnService
}

func NewBoardColumnHandler(boardService service.BoardService, boardColumnService service.BoardColumnService) *BoardColumnHandler {
	return &BoardColumnHandler{boardService: boardService, boardColumnService: boardColumnService}
}

func RegisterBoardColumnRoutes(router *mux.Router, handler *BoardColumnHandler) {
	router.HandleFunc("/boards/{board_id}/columns", handler.Create).Methods(http.MethodPost)
	router.HandleFunc("/boards/{board_id}/columns/order", handler.Reorder).Methods(http.MethodPut)
	router.HandleFunc("/boards/{board_id}/columns/{column_id}", handler.Update).Methods(http.MethodPut)
	router.HandleFunc("/boards/{board_id}/columns/{column_id}", handler.Delete).Methods(http.MethodDelete)
}

func (handler *BoardColumnHandler) Reorder(writer http.ResponseWriter, request *http.Request) {
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
	dtoValue := &dto.BoardColumnOrderDto{}
	if err := json.NewDecoder(request.Body).Decode(dtoValue); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	if err := handler.boardColumnService.Reorder(request.Context(), boardId, dtoValue.ColumnIds); err != nil {
		writeBoardError(writer, err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (handler *BoardColumnHandler) Create(writer http.ResponseWriter, request *http.Request) {
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
	column, err := handler.boardColumnService.Create(request.Context(), &model.BoardColumn{
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

func (handler *BoardColumnHandler) Update(writer http.ResponseWriter, request *http.Request) {
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
	column, err := handler.boardColumnService.Update(request.Context(), &model.BoardColumn{
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

func (handler *BoardColumnHandler) Delete(writer http.ResponseWriter, request *http.Request) {
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
	if err := handler.boardColumnService.Delete(request.Context(), boardId, columnId, targetColumnId); err != nil {
		writeBoardError(writer, err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (handler *BoardColumnHandler) requireOwnedBoard(writer http.ResponseWriter, request *http.Request, boardID int64, userID int64) bool {
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
