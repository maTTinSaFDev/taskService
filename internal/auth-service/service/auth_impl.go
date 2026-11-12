package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"auth-service/model"
	"auth-service/repository"
)

type AuthServiceImpl struct {
	authRepository    repository.AuthRepository
	sessionRepository repository.SessionRepository
	hasher            PasswordHasher
	tokens            AccessTokenManager
	refreshTTL        time.Duration
	now               func() time.Time
	dummyHash         string
}

func NewAuthService(authRepository repository.AuthRepository, sessionRepository repository.SessionRepository, hasher PasswordHasher, tokens AccessTokenManager, refreshTTL time.Duration) (*AuthServiceImpl, error) {
	if authRepository == nil || sessionRepository == nil || hasher == nil || tokens == nil || refreshTTL <= 0 {
		return nil, errors.New("auth service dependencies and refresh TTL are required")
	}
	dummyHash, err := hasher.Hash("timing-only-password")
	if err != nil {
		return nil, err
	}
	return &AuthServiceImpl{
		authRepository:    authRepository,
		sessionRepository: sessionRepository,
		hasher:            hasher,
		tokens:            tokens,
		refreshTTL:        refreshTTL,
		now:               time.Now,
		dummyHash:         dummyHash,
	}, nil
}

func (service *AuthServiceImpl) Register(ctx context.Context, name string, email string, password string) (*model.Session, error) {
	name = strings.TrimSpace(name)
	email, err := normalizeEmail(email)
	if err != nil || utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 100 || !validPassword(password) {
		return nil, ErrInvalidInput
	}
	passwordHash, err := service.hasher.Hash(password)
	if err != nil {
		return nil, err
	}
	user, err := service.authRepository.CreateUser(ctx, name, email, passwordHash)
	if errors.Is(err, repository.ErrEmailExists) {
		return nil, ErrEmailExists
	}
	if err != nil {
		return nil, err
	}
	return service.createSession(ctx, user)
}

func (service *AuthServiceImpl) Login(ctx context.Context, email string, password string) (*model.Session, error) {
	email, normalizeErr := normalizeEmail(email)
	if normalizeErr != nil || password == "" {
		return nil, ErrInvalidCredentials
	}
	credential, err := service.authRepository.GetCredentialByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		_ = service.hasher.Compare(service.dummyHash, password)
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if err := service.hasher.Compare(credential.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}
	return service.createSession(ctx, &credential.User)
}

func (service *AuthServiceImpl) Refresh(ctx context.Context, currentRefreshToken string) (*model.Session, error) {
	if strings.TrimSpace(currentRefreshToken) == "" {
		return nil, ErrInvalidRefresh
	}
	replacementRaw, replacementHash, err := newRefreshToken()
	if err != nil {
		return nil, err
	}
	replacementExpiry := service.now().UTC().Add(service.refreshTTL)
	user, err := service.sessionRepository.Rotate(ctx, hashRefreshToken(currentRefreshToken), replacementHash, service.refreshTTL)
	if errors.Is(err, repository.ErrInvalidRefresh) || errors.Is(err, repository.ErrNotFound) {
		return nil, ErrInvalidRefresh
	}
	if err != nil {
		return nil, err
	}
	accessToken, accessExpiry, err := service.tokens.Issue(user.ID)
	if err != nil {
		return nil, err
	}
	return &model.Session{AccessToken: accessToken, AccessExpiry: accessExpiry, RefreshToken: replacementRaw, RefreshExpiry: replacementExpiry, User: *user}, nil
}

func (service *AuthServiceImpl) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return nil
	}
	return service.sessionRepository.Revoke(ctx, hashRefreshToken(refreshToken))
}

func (service *AuthServiceImpl) Me(ctx context.Context, accessToken string) (*model.User, error) {
	userID, err := service.tokens.Verify(accessToken)
	if err != nil {
		return nil, ErrUnauthorized
	}
	user, err := service.authRepository.GetUserByID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrUnauthorized
	}
	return user, err
}

func (service *AuthServiceImpl) createSession(ctx context.Context, user *model.User) (*model.Session, error) {
	refreshRaw, refreshHash, err := newRefreshToken()
	if err != nil {
		return nil, err
	}
	refreshExpiry := service.now().UTC().Add(service.refreshTTL)
	if err := service.sessionRepository.Create(ctx, user, refreshHash, service.refreshTTL); err != nil {
		return nil, err
	}
	accessToken, accessExpiry, err := service.tokens.Issue(user.ID)
	if err != nil {
		return nil, err
	}
	return &model.Session{AccessToken: accessToken, AccessExpiry: accessExpiry, RefreshToken: refreshRaw, RefreshExpiry: refreshExpiry, User: *user}, nil
}

func normalizeEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value || len(value) > 320 {
		return "", ErrInvalidInput
	}
	return value, nil
}

func validPassword(password string) bool {
	length := len([]byte(password))
	return length >= 12 && length <= 72
}
