package transport

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type accessVerifierStub struct {
	userID int64
	err    error
}

func (stub accessVerifierStub) Verify(string) (int64, error) {
	return stub.userID, stub.err
}

func TestRequireAuth(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		verifier   accessVerifierStub
		wantStatus int
	}{
		{name: "missing token", wantStatus: http.StatusUnauthorized},
		{name: "invalid token", header: "Bearer broken", verifier: accessVerifierStub{err: errors.New("invalid")}, wantStatus: http.StatusUnauthorized},
		{name: "valid token", header: "Bearer valid", verifier: accessVerifierStub{userID: 42}, wantStatus: http.StatusNoContent},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			handler := RequireAuth(testCase.verifier)(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				userID, ok := authenticatedUserID(request)
				if !ok || userID != 42 {
					t.Fatalf("authenticated user = %d, %v", userID, ok)
				}
				writer.WriteHeader(http.StatusNoContent)
			}))
			request := httptest.NewRequest(http.MethodGet, "/tasks", nil)
			request.Header.Set("Authorization", testCase.header)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != testCase.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, testCase.wantStatus)
			}
		})
	}
}
