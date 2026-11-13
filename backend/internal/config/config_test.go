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
