package config

import (
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setValidEnv(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", "test")
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://app:app@localhost:5432/app?sslmode=disable")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,https://example.com")
	t.Setenv("TRUSTED_PROXY_CIDRS", "")
	t.Setenv("RATE_LIMIT_REQUESTS", "60")
	t.Setenv("RATE_LIMIT_WINDOW", "1m")
	t.Setenv("SESSION_IDLE_TTL", "720h")
	t.Setenv("SESSION_ABSOLUTE_TTL", "2160h")
	t.Setenv("LOGIN_MAX_FAILURES", "10")
	t.Setenv("LOGIN_FAILURE_WINDOW", "15m")
	t.Setenv("BOOTSTRAP_ADMIN_EMAIL", "")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_SERVICE_NAME", "sso-api")
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "1.0")
	t.Setenv("HTTP_MAX_BODY_BYTES", "1048576")
}

func TestLoadValid(t *testing.T) {
	setValidEnv(t)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, 60, cfg.RateLimitRequests)
	assert.Equal(t, time.Minute, cfg.RateLimitWindow)
	assert.Equal(t, []string{"http://localhost:5173", "https://example.com"}, cfg.CORSAllowedOrigins)
	assert.Equal(t, 720*time.Hour, cfg.SessionIdleTTL)
	assert.Equal(t, 2160*time.Hour, cfg.SessionAbsoluteTTL)
	assert.Equal(t, 10, cfg.LoginMaxFailures)
	assert.Equal(t, 15*time.Minute, cfg.LoginFailureWindow)
}

func TestLoadRejectsIdleSessionLongerThanAbsolute(t *testing.T) {
	setValidEnv(t)
	t.Setenv("SESSION_IDLE_TTL", "48h")
	t.Setenv("SESSION_ABSOLUTE_TTL", "24h")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SESSION_IDLE_TTL")
}

func TestLoadRejectsBootstrapPasswordWithoutEmail(t *testing.T) {
	setValidEnv(t)
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "correct horse battery staple")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "BOOTSTRAP_ADMIN_EMAIL")
}

func TestLoadAllowsBootstrapEmailWithoutPasswordForLaterRestarts(t *testing.T) {
	setValidEnv(t)
	t.Setenv("BOOTSTRAP_ADMIN_EMAIL", "admin@example.com")

	_, err := Load()
	require.NoError(t, err)
}

func TestLoadRejectsInvalidBootstrapEmail(t *testing.T) {
	setValidEnv(t)
	t.Setenv("BOOTSTRAP_ADMIN_EMAIL", "not-an-email")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "BOOTSTRAP_ADMIN_EMAIL")
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	setValidEnv(t)
	t.Setenv("DATABASE_URL", "")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL")
}

func TestLoadRejectsBadPort(t *testing.T) {
	setValidEnv(t)
	t.Setenv("PORT", "eighty")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PORT")
}

func TestLoadRejectsBadLogLevel(t *testing.T) {
	setValidEnv(t)
	t.Setenv("LOG_LEVEL", "loud")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "LOG_LEVEL")
}

func TestLoadRejectsBadSampleRatio(t *testing.T) {
	setValidEnv(t)
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "2")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OTEL_TRACES_SAMPLER_ARG")
}

func TestLoadRejectsRelativeOrigin(t *testing.T) {
	setValidEnv(t)
	t.Setenv("CORS_ALLOWED_ORIGINS", "/relative")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CORS_ALLOWED_ORIGINS")
}

func TestLoadRejectsBadTrustedProxyCIDR(t *testing.T) {
	setValidEnv(t)
	t.Setenv("TRUSTED_PROXY_CIDRS", "not-a-cidr")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TRUSTED_PROXY_CIDRS")
}

func TestLoadParsesTrustedProxyCIDRs(t *testing.T) {
	setValidEnv(t)
	t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/8,192.168.0.0/16")

	cfg, err := Load()
	require.NoError(t, err)
	require.Len(t, cfg.TrustedProxies, 2)
	assert.True(t, cfg.TrustedProxies[0].Contains(netip.MustParseAddr("10.1.2.3")))
	assert.True(t, cfg.TrustedProxies[1].Contains(netip.MustParseAddr("192.168.1.1")))
}
