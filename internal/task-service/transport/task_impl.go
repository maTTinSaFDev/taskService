package transport

import (
	"core/dto"
	"core/model"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"task-service/service"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskHandlerImpl struct {
	taskService service.TaskService
}

func NewTaskHandler(taskService service.TaskService) *TaskHandlerImpl {
	return &TaskHandlerImpl{taskService: taskService}
}

func (handler *TaskHandlerImpl) GetById(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}
	idString := mux.Vars(r)["id"]
	if idString == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("id is empty"))
		return
	}
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if id <= 0 {
		http.Error(w, "id must be greater than zero", http.StatusBadRequest)
		return
	}
	task, err := handler.taskService.GetById(r.Context(), id)
	if err != nil {
		writeGetByIDError(w, err)
		return
	}
	if task.UserId != userID {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func writeGetByIDError(w http.ResponseWriter, err error) {
	if errors.Is(err, model.ErrTaskNotFound) || status.Code(err) == codes.NotFound {
		http.Error(w, model.ErrTaskNotFound.Error(), http.StatusNotFound)
		return
	}

	switch status.Code(err) {
	case codes.InvalidArgument:
		http.Error(w, "invalid request", http.StatusBadRequest)
	case codes.Unavailable:
		http.Error(w, "database service unavailable", http.StatusBadGateway)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (handler *TaskHandlerImpl) GetByFilters(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}
	filters, err := taskFiltersFromRequest(r, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tasks, err := handler.taskService.GetByFilters(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, tasks)
}

func (handler *TaskHandlerImpl) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}
	taskDto := &dto.TaskCreateDto{}
	err := json.NewDecoder(r.Body).Decode(taskDto)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	taskModel := &model.Task{
		UserId:      userID,
		Title:       taskDto.Title,
		Description: taskDto.Description,
		ColumnId:    taskDto.ColumnId,
		Position:    taskDto.Position,
		Progress:    taskDto.Progress,
		IsDone:      taskDto.IsDone,
		StartTime:   taskDto.StartTime,
		Duration:    taskDto.Duration,
		Priority:    taskDto.Priority,
	}
	task, err := handler.taskService.Create(r.Context(), taskModel)

	if err != nil {
		// TODO: Once service errors are defined, map validation errors to 400 and internal failures to 500.
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (handler *TaskHandlerImpl) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}
	taskDto := &dto.TaskUpdateDto{}
	err := json.NewDecoder(r.Body).Decode(taskDto)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if taskDto.Id <= 0 {
		http.Error(w, "id must be greater than zero", http.StatusBadRequest)
		return
	}
	existing, err := handler.taskService.GetById(r.Context(), taskDto.Id)
	if err != nil {
		writeGetByIDError(w, err)
		return
	}
	if existing.UserId != userID {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	taskModel := &model.Task{
		Id:          taskDto.Id,
		UserId:      userID,
		Title:       taskDto.Title,
		Description: taskDto.Description,
		ColumnId:    taskDto.ColumnId,
		Position:    taskDto.Position,
		Progress:    taskDto.Progress,
		IsDone:      taskDto.IsDone,
		StartTime:   taskDto.StartTime,
		Duration:    taskDto.Duration,
		Priority:    taskDto.Priority,
	}
	err = handler.taskService.Update(r.Context(), taskModel)
	if err != nil {
		// TODO: Once service errors are defined, map validation errors to 400, not found to 404, and internal failures to 500.
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
}

func (handler *TaskHandlerImpl) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}
	idString := mux.Vars(r)["id"]
	if idString == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("id is empty"))
		return
	}
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if id <= 0 {
		http.Error(w, "id must be greater than zero", http.StatusBadRequest)
		return
	}
	existing, err := handler.taskService.GetById(r.Context(), id)
	if err != nil {
		writeGetByIDError(w, err)
		return
	}
	if existing.UserId != userID {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	err = handler.taskService.Delete(r.Context(), id)
	if err != nil {
		// TODO: Once service errors are defined, map not found to 404 and internal failures to 500.
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
