package repository

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// loginFailureWindow increments a counter, and restarts the window once the
// failure limit is reached so a lockout always lasts the full window.
var loginFailureWindow = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 or tonumber(current) >= tonumber(ARGV[2]) then
	redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return current
`)

type memoryFailure struct {
	count     int
	expiresAt time.Time
}

// maxMemoryFailures bounds the fallback map during a Redis outage so a spray of
// unique emails cannot grow process memory without limit.
const maxMemoryFailures = 10_000

// LoginAttempts bounds password guessing per email and IP. Redis is the
// primary counter; the in-process map takes over when Redis is unavailable so
// brute force stays bounded during an outage.
type LoginAttempts struct {
	redis  *redis.Client
	limit  int
	window time.Duration

	mu       sync.Mutex
	failures map[string]memoryFailure
	now      func() time.Time
}

// NewLoginAttempts builds the login limiter.
func NewLoginAttempts(client *redis.Client, limit int, window time.Duration) *LoginAttempts {
	return &LoginAttempts{
		redis:    client,
		limit:    limit,
		window:   window,
		failures: map[string]memoryFailure{},
		now:      time.Now,
	}
}

// Blocked reports the remaining lockout for an email and IP pair.
func (l *LoginAttempts) Blocked(ctx context.Context, email, ip string) (time.Duration, error) {
	key := loginAttemptKey(email, ip)

	redisBlocked, redisErr := l.redisBlocked(ctx, key)
	memoryBlocked := l.memoryBlocked(key)
	if redisBlocked > memoryBlocked {
		return redisBlocked, nil
	}
	if memoryBlocked > 0 {
		return memoryBlocked, nil
	}
	if redisErr != nil {
		return 0, fmt.Errorf("read login attempts: %w", redisErr)
	}
	return 0, nil
}

// RecordFailure increments the counter for an email and IP pair.
func (l *LoginAttempts) RecordFailure(ctx context.Context, email, ip string) error {
	key := loginAttemptKey(email, ip)

	timeoutCtx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()

	err := loginFailureWindow.Run(timeoutCtx, l.redis, []string{key}, l.window.Milliseconds(), l.limit).Err()
	if err != nil {
		l.recordMemoryFailure(key)
		return nil
	}
	return nil
}

// Reset clears the counter after a successful login.
func (l *LoginAttempts) Reset(ctx context.Context, email, ip string) error {
	key := loginAttemptKey(email, ip)

	l.mu.Lock()
	delete(l.failures, key)
	l.mu.Unlock()

	timeoutCtx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	if err := l.redis.Del(timeoutCtx, key).Err(); err != nil {
		return fmt.Errorf("clear login attempts: %w", err)
	}
	return nil
}

func (l *LoginAttempts) redisBlocked(ctx context.Context, key string) (time.Duration, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()

	count, err := l.redis.Get(timeoutCtx, key).Int()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if count < l.limit {
		return 0, nil
	}

	ttlCtx, cancelTTL := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancelTTL()
	ttl, err := l.redis.PTTL(ttlCtx, key).Result()
	if err != nil {
		return l.window, nil
	}
	if ttl <= 0 {
		return l.window, nil
	}
	return ttl, nil
}

func (l *LoginAttempts) memoryBlocked(key string) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.failures[key]
	if !ok {
		return 0
	}
	now := l.now()
	if !now.Before(entry.expiresAt) {
		delete(l.failures, key)
		return 0
	}
	if entry.count < l.limit {
		return 0
	}
	return entry.expiresAt.Sub(now)
}

func (l *LoginAttempts) recordMemoryFailure(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	entry, ok := l.failures[key]
	if !ok {
		if len(l.failures) >= maxMemoryFailures {
			l.makeRoomMemoryLocked(now)
		}
	}
	if !ok || !now.Before(entry.expiresAt) {
		entry = memoryFailure{expiresAt: now.Add(l.window)}
	}
	entry.count++
	if entry.count == 1 || entry.count >= l.limit {
		entry.expiresAt = now.Add(l.window)
	}
	l.failures[key] = entry
}

func (l *LoginAttempts) makeRoomMemoryLocked(now time.Time) {
	for key, entry := range l.failures {
		if !now.Before(entry.expiresAt) {
			delete(l.failures, key)
		}
	}
	if len(l.failures) < maxMemoryFailures {
		return
	}
	for key := range l.failures {
		delete(l.failures, key)
		break
	}
}

func loginAttemptKey(email, ip string) string {
	hash := sha256.Sum256([]byte(email))
	return fmt.Sprintf("login:fail:%x:%s", hash, ip)
}
