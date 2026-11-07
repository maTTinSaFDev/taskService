package service

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"
)

func TestJWTManagerIssueAndVerify(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	manager, err := NewJWTManager(privateKey, "issuer", "audience", 15*time.Minute)
	if err != nil {
		t.Fatalf("NewJWTManager returned error: %v", err)
	}
	now := time.Date(2026, time.November, 7, 12, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return now }
	token, expiry, err := manager.Issue(42)
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}
	if expiry != now.Add(15*time.Minute) {
		t.Fatalf("expiry = %v", expiry)
	}
	userID, err := manager.Verify(token)
	if err != nil || userID != 42 {
		t.Fatalf("Verify = (%d, %v), want (42, nil)", userID, err)
	}
	manager.now = func() time.Time { return expiry.Add(time.Second) }
	if _, err := manager.Verify(token); err == nil {
		t.Fatal("expired access token was accepted")
	}
}
