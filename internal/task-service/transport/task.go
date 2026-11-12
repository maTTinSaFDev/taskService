package transport

import "net/http"

type TaskHandler interface {
	GetById(w http.ResponseWriter, r *http.Request)
	GetByFilters(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}
