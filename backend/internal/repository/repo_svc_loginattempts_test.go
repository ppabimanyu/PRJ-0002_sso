package repository

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return server, client
}

func TestLoginAttemptsBlocksAfterTheLimit(t *testing.T) {
	ctx := context.Background()
	_, client := newTestRedis(t)
	attempts := NewLoginAttempts(client, 3, 15*time.Minute)

	for range 3 {
		require.NoError(t, attempts.RecordFailure(ctx, "person@example.com", "203.0.113.10"))
	}

	blocked, err := attempts.Blocked(ctx, "person@example.com", "203.0.113.10")
	require.NoError(t, err)
	assert.Positive(t, blocked)

	otherIP, err := attempts.Blocked(ctx, "person@example.com", "203.0.113.11")
	require.NoError(t, err)
	assert.Zero(t, otherIP)

	require.NoError(t, attempts.Reset(ctx, "person@example.com", "203.0.113.10"))
	blocked, err = attempts.Blocked(ctx, "person@example.com", "203.0.113.10")
	require.NoError(t, err)
	assert.Zero(t, blocked)
}

func TestLoginAttemptsExpireWithTheWindow(t *testing.T) {
	ctx := context.Background()
	server, client := newTestRedis(t)
	attempts := NewLoginAttempts(client, 3, 15*time.Minute)

	for range 3 {
		require.NoError(t, attempts.RecordFailure(ctx, "person@example.com", "203.0.113.10"))
	}
	server.FastForward(16 * time.Minute)

	blocked, err := attempts.Blocked(ctx, "person@example.com", "203.0.113.10")
	require.NoError(t, err)
	assert.Zero(t, blocked)
}

func TestLoginAttemptsFallBackToMemoryWhenRedisIsDown(t *testing.T) {
	ctx := context.Background()
	_, client := newTestRedis(t)
	attempts := NewLoginAttempts(client, 3, 15*time.Minute)
	require.NoError(t, client.Close())

	for range 3 {
		require.NoError(t, attempts.RecordFailure(ctx, "person@example.com", "203.0.113.10"))
	}

	blocked, err := attempts.Blocked(ctx, "person@example.com", "203.0.113.10")
	require.NoError(t, err)
	assert.Positive(t, blocked)
}

func TestLoginAttemptsMemoryWindowExpires(t *testing.T) {
	ctx := context.Background()
	_, client := newTestRedis(t)
	attempts := NewLoginAttempts(client, 3, 15*time.Minute)
	now := time.Now()
	attempts.now = func() time.Time { return now }
	require.NoError(t, client.Close())

	for range 3 {
		require.NoError(t, attempts.RecordFailure(ctx, "person@example.com", "203.0.113.10"))
	}
	blocked, err := attempts.Blocked(ctx, "person@example.com", "203.0.113.10")
	require.NoError(t, err)
	assert.Positive(t, blocked)

	now = now.Add(16 * time.Minute)
	blocked, err = attempts.Blocked(ctx, "person@example.com", "203.0.113.10")
	assert.Zero(t, blocked)
	assert.Error(t, err)
}
