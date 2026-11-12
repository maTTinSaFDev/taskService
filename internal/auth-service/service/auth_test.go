package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"auth-service/model"
)

type authRepositoryStub struct {
	createUserFunc    func(context.Context, string, string, string) (*model.User, error)
	getCredentialFunc func(context.Context, string) (*model.Credential, error)
	getUserFunc       func(context.Context, int64) (*model.User, error)
}

func (stub *authRepositoryStub) CreateUser(ctx context.Context, name, email, hash string) (*model.User, error) {
	return stub.createUserFunc(ctx, name, email, hash)
}
func (stub *authRepositoryStub) GetCredentialByEmail(ctx context.Context, email string) (*model.Credential, error) {
	return stub.getCredentialFunc(ctx, email)
}
func (stub *authRepositoryStub) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
	return stub.getUserFunc(ctx, userID)
}

type sessionRepositoryStub struct {
	createFunc func(context.Context, *model.User, string, time.Duration) error
	rotateFunc func(context.Context, string, string, time.Duration) (*model.User, error)
	revokeFunc func(context.Context, string) error
}

func (stub *sessionRepositoryStub) Create(ctx context.Context, user *model.User, hash string, ttl time.Duration) error {
	return stub.createFunc(ctx, user, hash, ttl)
}
func (stub *sessionRepositoryStub) Rotate(ctx context.Context, current, replacement string, ttl time.Duration) (*model.User, error) {
	return stub.rotateFunc(ctx, current, replacement, ttl)
}
func (stub *sessionRepositoryStub) Revoke(ctx context.Context, hash string) error {
	return stub.revokeFunc(ctx, hash)
}

type passwordHasherStub struct{}

func (passwordHasherStub) Hash(password string) (string, error)       { return "hashed:" + password, nil }
func (passwordHasherStub) Compare(hash string, password string) error { return nil }

type rejectingPasswordHasherStub struct{}

func (rejectingPasswordHasherStub) Hash(password string) (string, error) { return "dummy", nil }
func (rejectingPasswordHasherStub) Compare(hash string, password string) error {
	return errors.New("password mismatch")
}

type tokenManagerStub struct{}

func (tokenManagerStub) Issue(userID int64) (string, time.Time, error) {
	return "access-token", time.Unix(2000, 0), nil
}
func (tokenManagerStub) Verify(string) (int64, error) { return 42, nil }

func TestRegisterNormalizesEmailAndPersistsOnlyHashes(t *testing.T) {
	var storedRefreshHash string
	authRepository := &authRepositoryStub{
		createUserFunc: func(_ context.Context, name, email, passwordHash string) (*model.User, error) {
			if name != "Vadim" || email != "user@example.com" || passwordHash != "hashed:very-secure-password" {
				t.Fatalf("user input = %q, %q, %q", name, email, passwordHash)
			}
			return &model.User{ID: 42, Name: name, Email: email}, nil
		},
	}
	sessionRepository := &sessionRepositoryStub{
		createFunc: func(_ context.Context, user *model.User, tokenHash string, ttl time.Duration) error {
			if user.ID != 42 || len(tokenHash) != 64 || ttl != 24*time.Hour {
				t.Fatalf("refresh persistence = user %d, hash %q, ttl %v", user.ID, tokenHash, ttl)
			}
			storedRefreshHash = tokenHash
			return nil
		},
	}
	authService, err := NewAuthService(authRepository, sessionRepository, passwordHasherStub{}, tokenManagerStub{}, 24*time.Hour)
	if err != nil {
		t.Fatalf("NewAuthService returned error: %v", err)
	}
	session, err := authService.Register(context.Background(), " Vadim ", " USER@Example.com ", "very-secure-password")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if session.RefreshToken == "" || session.RefreshToken == storedRefreshHash || hashRefreshToken(session.RefreshToken) != storedRefreshHash {
		t.Fatal("refresh token was not separated from its persisted hash")
	}
}

func TestRefreshRotatesHashedToken(t *testing.T) {
	const currentRaw = "current-refresh-secret"
	var replacementHash string
	sessionRepository := &sessionRepositoryStub{
		rotateFunc: func(_ context.Context, currentHash string, nextHash string, ttl time.Duration) (*model.User, error) {
			if currentHash != hashRefreshToken(currentRaw) || currentHash == currentRaw {
				t.Fatalf("current hash = %q", currentHash)
			}
			if len(nextHash) != 64 || nextHash == currentHash || ttl != 24*time.Hour {
				t.Fatalf("replacement hash = %q", nextHash)
			}
			replacementHash = nextHash
			return &model.User{ID: 42, Name: "Vadim", Email: "user@example.com"}, nil
		},
	}
	authService, err := NewAuthService(&authRepositoryStub{}, sessionRepository, passwordHasherStub{}, tokenManagerStub{}, 24*time.Hour)
	if err != nil {
		t.Fatalf("NewAuthService returned error: %v", err)
	}
	session, err := authService.Refresh(context.Background(), currentRaw)
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if session.RefreshToken == currentRaw || session.RefreshToken == replacementHash || hashRefreshToken(session.RefreshToken) != replacementHash {
		t.Fatal("refresh session was not rotated correctly")
	}
}

func TestLoginReturnsGenericErrorForWrongPassword(t *testing.T) {
	repository := &authRepositoryStub{
		getCredentialFunc: func(context.Context, string) (*model.Credential, error) {
			return &model.Credential{
				User:         model.User{ID: 42, Name: "Vadim", Email: "user@example.com"},
				PasswordHash: "stored-hash",
			}, nil
		},
	}
	authService, err := NewAuthService(repository, &sessionRepositoryStub{}, rejectingPasswordHasherStub{}, tokenManagerStub{}, 24*time.Hour)
	if err != nil {
		t.Fatalf("NewAuthService returned error: %v", err)
	}
	if _, err := authService.Login(context.Background(), "user@example.com", "wrong-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestLogoutRevokesOnlyHashedRefreshToken(t *testing.T) {
	const raw = "refresh-token-to-revoke"
	sessionRepository := &sessionRepositoryStub{
		revokeFunc: func(_ context.Context, tokenHash string) error {
			if tokenHash == raw || tokenHash != hashRefreshToken(raw) {
				t.Fatalf("revoke token hash = %q", tokenHash)
			}
			return nil
		},
	}
	authService, err := NewAuthService(&authRepositoryStub{}, sessionRepository, passwordHasherStub{}, tokenManagerStub{}, 24*time.Hour)
	if err != nil {
		t.Fatalf("NewAuthService returned error: %v", err)
	}
	if err := authService.Logout(context.Background(), raw); err != nil {
		t.Fatalf("Logout returned error: %v", err)
	}
}
