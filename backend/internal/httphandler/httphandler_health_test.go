package httphandler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ppabimanyu/sso/internal/domain"
)

type stubChecker struct {
	err error
}

func (s stubChecker) Ping(context.Context) error {
	return s.err
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestLiveReturnsOK(t *testing.T) {
	handler := NewHealthHandler(nil, testLogger())
	recorder := httptest.NewRecorder()

	handler.Live(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	var body HealthResponse
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&body))
	assert.Equal(t, "ok", body.Status)
}

func TestReadyReturnsOKWhenEveryCheckPasses(t *testing.T) {
	handler := NewHealthHandler(map[string]domain.HealthChecker{
		"postgres": stubChecker{},
		"redis":    stubChecker{},
	}, testLogger())
	recorder := httptest.NewRecorder()

	handler.Ready(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	var body HealthResponse
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&body))
	assert.Equal(t, "ok", body.Status)
	assert.Equal(t, "ok", body.Checks["postgres"])
	assert.Equal(t, "ok", body.Checks["redis"])
}

func TestReadyReturns503WhenACheckFails(t *testing.T) {
	handler := NewHealthHandler(map[string]domain.HealthChecker{
		"postgres": stubChecker{},
		"redis":    stubChecker{err: errors.New("connection refused")},
	}, testLogger())
	recorder := httptest.NewRecorder()

	handler.Ready(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	var body HealthResponse
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&body))
	assert.Equal(t, "degraded", body.Status)
	assert.Equal(t, "error", body.Checks["redis"])
}

func TestRouterReturnsProblemJSONForUnknownRoute(t *testing.T) {
	router := NewRouter(RouterOptions{
		Logger:       testLogger(),
		MaxBodyBytes: 1 << 20,
	})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/missing", nil))

	require.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Equal(t, "application/problem+json", recorder.Header().Get("Content-Type"))
	assert.Contains(t, recorder.Body.String(), `"status":404`)
}
