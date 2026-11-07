package transport

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type AccessTokenVerifier interface {
	Verify(token string) (int64, error)
}

type authenticatedUserKey struct{}

func RequireAuth(verifier AccessTokenVerifier) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			parts := strings.Fields(request.Header.Get("Authorization"))
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(writer, "unauthorized", http.StatusUnauthorized)
				return
			}
			userID, err := verifier.Verify(parts[1])
			if err != nil {
				http.Error(writer, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(writer, request.WithContext(withAuthenticatedUser(request.Context(), userID)))
		})
	}
}

func withAuthenticatedUser(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, authenticatedUserKey{}, userID)
}

func authenticatedUserID(request *http.Request) (int64, bool) {
	userID, ok := request.Context().Value(authenticatedUserKey{}).(int64)
	return userID, ok && userID > 0
}

func requireAuthenticatedUser(writer http.ResponseWriter, request *http.Request) (int64, bool) {
	userID, ok := authenticatedUserID(request)
	if !ok {
		http.Error(writer, "unauthorized", http.StatusUnauthorized)
	}
	return userID, ok
}
