package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ppabimanyu/sso/internal/pkg/problem"
)

// fixedWindow increments a counter and gives it a TTL on first use, in one
// atomic round trip.
var fixedWindow = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
	redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return current
`)

// RateLimit builds a fixed-window limiter keyed by client IP and request path.
// Redis errors fail open so an outage does not take the API down with it.
func RateLimit(client *redis.Client, limit int, window time.Duration, logger *slog.Logger) func(http.Handler) http.Handler {
	windowMillis := window.Milliseconds()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 250*time.Millisecond)
			defer cancel()

			key := fmt.Sprintf("ratelimit:%s:%s", ClientIP(r), r.URL.Path)
			count, err := fixedWindow.Run(ctx, client, []string{key}, windowMillis).Int64()
			if err != nil {
				logger.WarnContext(r.Context(), "rate limit unavailable, failing open", "error", err)
				next.ServeHTTP(w, r)
				return
			}

			remaining := limit - int(count)
			if remaining < 0 {
				remaining = 0
			}
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			if int(count) > limit {
				w.Header().Set("Retry-After", strconv.Itoa(int(window.Seconds())))
				problem.WriteStatus(w, r, problem.New(http.StatusTooManyRequests, "Too Many Requests", "Rate limit exceeded. Retry later."))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ClientIP returns the request peer address without its port. RealIP has
// already replaced the peer when a trusted proxy supplied X-Forwarded-For.
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
