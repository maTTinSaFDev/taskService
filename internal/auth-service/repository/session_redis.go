package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"auth-service/model"
	"github.com/redis/go-redis/v9"
)

const refreshSessionKeyPrefix = "refresh_session:"

var rotateRefreshSessionScript = redis.NewScript(`
local current = redis.call("GET", KEYS[1])
if not current then
    return nil
end
if redis.call("EXISTS", KEYS[2]) == 1 then
    return redis.error_reply("replacement refresh session already exists")
end
redis.call("SET", KEYS[2], current, "PX", ARGV[1])
redis.call("DEL", KEYS[1])
return current
`)

type RedisSessionRepository struct {
	client redis.UniversalClient
}

func NewRedisSessionRepository(client redis.UniversalClient) SessionRepository {
	return &RedisSessionRepository{client: client}
}

func (repository *RedisSessionRepository) Create(ctx context.Context, user *model.User, tokenHash string, ttl time.Duration) error {
	if user == nil || user.ID <= 0 || tokenHash == "" || ttl <= 0 {
		return errors.New("valid refresh session is required")
	}
	value, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("encode refresh session: %w", err)
	}
	return repository.client.Set(ctx, refreshSessionKey(tokenHash), value, ttl).Err()
}

func (repository *RedisSessionRepository) Rotate(ctx context.Context, currentTokenHash string, replacementTokenHash string, ttl time.Duration) (*model.User, error) {
	if currentTokenHash == "" || replacementTokenHash == "" || ttl <= 0 {
		return nil, ErrInvalidRefresh
	}
	value, err := rotateRefreshSessionScript.Run(
		ctx,
		repository.client,
		[]string{refreshSessionKey(currentTokenHash), refreshSessionKey(replacementTokenHash)},
		ttl.Milliseconds(),
	).Text()
	if errors.Is(err, redis.Nil) {
		return nil, ErrInvalidRefresh
	}
	if err != nil {
		return nil, err
	}
	user := &model.User{}
	if err := json.Unmarshal([]byte(value), user); err != nil {
		return nil, fmt.Errorf("decode refresh session: %w", err)
	}
	if user.ID <= 0 {
		return nil, ErrInvalidRefresh
	}
	return user, nil
}

func (repository *RedisSessionRepository) Revoke(ctx context.Context, tokenHash string) error {
	if tokenHash == "" {
		return nil
	}
	return repository.client.Del(ctx, refreshSessionKey(tokenHash)).Err()
}

func refreshSessionKey(tokenHash string) string {
	return refreshSessionKeyPrefix + tokenHash
}
