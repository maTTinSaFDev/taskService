package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTVerifier(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	verifier, err := NewJWTVerifier(publicKey, "taskflow-auth", "taskflow-api")
	if err != nil {
		t.Fatalf("NewJWTVerifier returned error: %v", err)
	}
	now := time.Date(2026, time.November, 7, 12, 0, 0, 0, time.UTC)
	verifier.now = func() time.Time { return now }
	claims := jwt.RegisteredClaims{
		Issuer: "taskflow-auth", Subject: strconv.FormatInt(42, 10),
		Audience: jwt.ClaimStrings{"taskflow-api"},
		IssuedAt: jwt.NewNumericDate(now), NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
	}
	value, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	userID, err := verifier.Verify(value)
	if err != nil || userID != 42 {
		t.Fatalf("Verify = (%d, %v), want (42, nil)", userID, err)
	}
	verifier.now = func() time.Time { return now.Add(2 * time.Minute) }
	if _, err := verifier.Verify(value); err == nil {
		t.Fatal("expired token was accepted")
	}
}
