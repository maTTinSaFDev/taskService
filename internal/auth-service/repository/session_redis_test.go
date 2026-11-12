package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"auth-service/model"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisSessionRepositoryLifecycle(t *testing.T) {
	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	repository := NewRedisSessionRepository(client)
	ctx := context.Background()
	user := &model.User{ID: 42, Name: "Vadim", Email: "user@example.com"}

	if err := repository.Create(ctx, user, "current-hash", time.Hour); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if ttl := redisServer.TTL(refreshSessionKey("current-hash")); ttl != time.Hour {
		t.Fatalf("created session TTL = %v, want %v", ttl, time.Hour)
	}

	rotatedUser, err := repository.Rotate(ctx, "current-hash", "replacement-hash", 30*time.Minute)
	if err != nil {
		t.Fatalf("Rotate returned error: %v", err)
	}
	if *rotatedUser != *user {
		t.Fatalf("rotated user = %+v, want %+v", rotatedUser, user)
	}
	if redisServer.Exists(refreshSessionKey("current-hash")) {
		t.Fatal("current refresh session still exists after rotation")
	}
	if ttl := redisServer.TTL(refreshSessionKey("replacement-hash")); ttl != 30*time.Minute {
		t.Fatalf("replacement session TTL = %v, want %v", ttl, 30*time.Minute)
	}

	if _, err := repository.Rotate(ctx, "current-hash", "another-hash", time.Hour); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("reusing rotated session returned %v, want %v", err, ErrInvalidRefresh)
	}
	if err := repository.Revoke(ctx, "replacement-hash"); err != nil {
		t.Fatalf("Revoke returned error: %v", err)
	}
	if redisServer.Exists(refreshSessionKey("replacement-hash")) {
		t.Fatal("refresh session still exists after revoke")
	}
}
