package model

import (
	"errors"
	"time"
)

var (
	ErrEmailAlreadyExists     = errors.New("email already exists")
	ErrCredentialNotFound     = errors.New("credential not found")
	ErrRefreshSessionNotFound = errors.New("refresh session not found")
	ErrRefreshSessionInvalid  = errors.New("refresh session is expired or revoked")
)

// AuthIdentity contains only profile data that may cross service boundaries.
type AuthIdentity struct {
	UserID int64
	Name   string
	Email  string
}

// AuthCredential is an internal persistence record. PasswordHash must never be
// serialized by an HTTP transport.
type AuthCredential struct {
	Identity     AuthIdentity
	PasswordHash string `json:"-"`
}

type RefreshSession struct {
	UserID    int64
	TokenHash string `json:"-"`
	ExpiresAt time.Time
}
