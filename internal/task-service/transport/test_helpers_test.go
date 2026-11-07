package transport

import "net/http"

func authenticatedRequest(request *http.Request, userID int64) *http.Request {
	return request.WithContext(withAuthenticatedUser(request.Context(), userID))
}
