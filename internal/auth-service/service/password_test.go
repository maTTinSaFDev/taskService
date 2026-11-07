package service

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestBcryptHasher(t *testing.T) {
	hasher := NewBcryptHasher(bcrypt.MinCost)
	hash, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash returned error: %v", err)
	}
	if hash == "correct horse battery staple" {
		t.Fatal("password was stored without hashing")
	}
	if err := hasher.Compare(hash, "correct horse battery staple"); err != nil {
		t.Fatalf("Compare rejected correct password: %v", err)
	}
	if err := hasher.Compare(hash, "wrong password"); err == nil {
		t.Fatal("Compare accepted wrong password")
	}
}
