package transport

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(handler TaskHandler) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/tasks/{id}", handler.GetById).Methods(http.MethodGet)
	router.HandleFunc("/tasks", handler.GetByFilters).Methods(http.MethodGet)
	router.HandleFunc("/tasks", handler.Create).Methods(http.MethodPost)
	router.HandleFunc("/tasks", handler.Update).Methods(http.MethodPut)
	router.HandleFunc("/tasks/{id}", handler.Delete).Methods(http.MethodDelete)

	return router
}
