package transport

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"auth-service/model"
	"auth-service/service"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const refreshCookieName = "taskflow_refresh"

type Handler struct {
	service      *service.AuthService
	cookieSecure bool
}

func NewRouter(authService *service.AuthService, cookieSecure bool) *mux.Router {
	handler := &Handler{service: authService, cookieSecure: cookieSecure}
	router := mux.NewRouter()
	router.HandleFunc("/healthz", handler.health).Methods(http.MethodGet)
	router.HandleFunc("/auth/register", handler.register).Methods(http.MethodPost)
	router.HandleFunc("/auth/login", handler.login).Methods(http.MethodPost)
	router.HandleFunc("/auth/refresh", handler.refresh).Methods(http.MethodPost)
	router.HandleFunc("/auth/logout", handler.logout).Methods(http.MethodPost)
	router.HandleFunc("/auth/me", handler.me).Methods(http.MethodGet)
	return router
}

type credentialsRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type sessionResponse struct {
	AccessToken          string     `json:"access_token"`
	AccessTokenExpiresAt time.Time  `json:"access_token_expires_at"`
	User                 model.User `json:"user"`
}

func (handler *Handler) register(writer http.ResponseWriter, request *http.Request) {
	var body credentialsRequest
	if err := decodeJSON(writer, request, &body); err != nil {
		return
	}
	session, err := handler.service.Register(request.Context(), body.Name, body.Email, body.Password)
	if err != nil {
		writeAuthError(writer, err)
		return
	}
	handler.writeSession(writer, http.StatusCreated, session)
}

func (handler *Handler) login(writer http.ResponseWriter, request *http.Request) {
	var body credentialsRequest
	if err := decodeJSON(writer, request, &body); err != nil {
		return
	}
	session, err := handler.service.Login(request.Context(), body.Email, body.Password)
	if err != nil {
		writeAuthError(writer, err)
		return
	}
	handler.writeSession(writer, http.StatusOK, session)
}

func (handler *Handler) refresh(writer http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie(refreshCookieName)
	if err != nil {
		writeAuthError(writer, service.ErrInvalidRefresh)
		return
	}
	session, err := handler.service.Refresh(request.Context(), cookie.Value)
	if err != nil {
		handler.clearRefreshCookie(writer)
		writeAuthError(writer, err)
		return
	}
	handler.writeSession(writer, http.StatusOK, session)
}

func (handler *Handler) logout(writer http.ResponseWriter, request *http.Request) {
	if cookie, err := request.Cookie(refreshCookieName); err == nil {
		if err := handler.service.Logout(request.Context(), cookie.Value); err != nil {
			writeAuthError(writer, err)
			return
		}
	}
	handler.clearRefreshCookie(writer)
	writer.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) me(writer http.ResponseWriter, request *http.Request) {
	token, ok := bearerToken(request.Header.Get("Authorization"))
	if !ok {
		writeAuthError(writer, service.ErrUnauthorized)
		return
	}
	user, err := handler.service.Me(request.Context(), token)
	if err != nil {
		writeAuthError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, user)
}

func (handler *Handler) health(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) writeSession(writer http.ResponseWriter, statusCode int, session *model.Session) {
	http.SetCookie(writer, &http.Cookie{
		Name: refreshCookieName, Value: session.RefreshToken, Path: "/auth",
		HttpOnly: true, Secure: handler.cookieSecure, SameSite: http.SameSiteLaxMode,
		Expires: session.RefreshExpiry, MaxAge: int(time.Until(session.RefreshExpiry).Seconds()),
	})
	writeJSON(writer, statusCode, sessionResponse{AccessToken: session.AccessToken, AccessTokenExpiresAt: session.AccessExpiry, User: session.User})
}

func (handler *Handler) clearRefreshCookie(writer http.ResponseWriter) {
	http.SetCookie(writer, &http.Cookie{
		Name: refreshCookieName, Value: "", Path: "/auth", HttpOnly: true,
		Secure: handler.cookieSecure, SameSite: http.SameSiteLaxMode, MaxAge: -1, Expires: time.Unix(1, 0),
	})
}

func decodeJSON(writer http.ResponseWriter, request *http.Request, destination any) error {
	request.Body = http.MaxBytesReader(writer, request.Body, 1<<20)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		http.Error(writer, "invalid JSON request", http.StatusBadRequest)
		return err
	}
	return nil
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	returnValue := ""
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		returnValue = parts[1]
	}
	return returnValue, returnValue != ""
}

func writeAuthError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		http.Error(writer, "name, valid email and password of 12-72 bytes are required", http.StatusBadRequest)
	case errors.Is(err, service.ErrEmailExists):
		http.Error(writer, err.Error(), http.StatusConflict)
	case errors.Is(err, service.ErrInvalidCredentials), errors.Is(err, service.ErrInvalidRefresh), errors.Is(err, service.ErrUnauthorized):
		http.Error(writer, "unauthorized", http.StatusUnauthorized)
	case status.Code(err) == codes.Unavailable:
		http.Error(writer, "database service unavailable", http.StatusServiceUnavailable)
	default:
		log.Printf("auth request failed: %v", err)
		http.Error(writer, "internal server error", http.StatusInternalServerError)
	}
}

func writeJSON(writer http.ResponseWriter, statusCode int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		log.Printf("write JSON response: %v", err)
	}
}
