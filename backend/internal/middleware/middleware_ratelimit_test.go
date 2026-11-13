package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestClient(t *testing.T, server *miniredis.Miniredis) *redis.Client {
	t.Helper()
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestRateLimitAllowsUntilLimitThenBlocks(t *testing.T) {
	server := miniredis.RunT(t)
	client := newTestClient(t, server)

	handler := RateLimit(client, 2, time.Minute, testLogger())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	request := func() *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/widgets", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		handler.ServeHTTP(recorder, req)
		return recorder
	}

	first := request()
	require.Equal(t, http.StatusOK, first.Code)
	assert.Equal(t, "2", first.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "1", first.Header().Get("X-RateLimit-Remaining"))

	second := request()
	require.Equal(t, http.StatusOK, second.Code)
	assert.Equal(t, "0", second.Header().Get("X-RateLimit-Remaining"))

	third := request()
	require.Equal(t, http.StatusTooManyRequests, third.Code)
	assert.Equal(t, "0", third.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, third.Header().Get("Retry-After"))
	assert.Contains(t, third.Body.String(), "Too Many Requests")
}

func TestRateLimitFailsOpenWhenRedisIsDown(t *testing.T) {
	server := miniredis.RunT(t)
	client := newTestClient(t, server)
	server.Close()

	called := false
	handler := RateLimit(client, 1, time.Minute, testLogger())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(recorder, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestRateLimitKeysByIPAndPath(t *testing.T) {
	server := miniredis.RunT(t)
	client := newTestClient(t, server)

	handler := RateLimit(client, 1, time.Minute, testLogger())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	do := func(addr, path string) int {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.RemoteAddr = addr
		handler.ServeHTTP(recorder, req)
		return recorder.Code
	}

	assert.Equal(t, http.StatusOK, do("203.0.113.10:1", "/a"))
	assert.Equal(t, http.StatusOK, do("203.0.113.11:1", "/a"))
	assert.Equal(t, http.StatusOK, do("203.0.113.10:1", "/b"))
	assert.Equal(t, http.StatusTooManyRequests, do("203.0.113.10:1", "/a"))
}
