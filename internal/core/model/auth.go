package model

import (
	"errors"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrCredentialNotFound = errors.New("credential not found")
)

// AuthCredential is an internal persistence record. PasswordHash must never be
// serialized by an HTTP transport.
type AuthCredential struct {
	User         User
	PasswordHash string `json:"-"`
}
