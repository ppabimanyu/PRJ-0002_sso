package repository

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisHealth implements domain.HealthChecker for Redis.
type RedisHealth struct {
	client *redis.Client
}

// NewRedisHealth builds the Redis health checker.
func NewRedisHealth(client *redis.Client) *RedisHealth {
	return &RedisHealth{client: client}
}

// Ping satisfies domain.HealthChecker for Redis.
func (h *RedisHealth) Ping(ctx context.Context) error {
	if err := h.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}
	return nil
}
