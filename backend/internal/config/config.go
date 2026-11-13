// Package config loads and validates environment settings.
package config

import (
	"fmt"
	"log/slog"
	"net/netip"
	"net/url"
	"strconv"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config holds every setting the service reads from the environment.
// Load parses and validates it once at startup.
type Config struct {
	AppEnv string `env:"APP_ENV" envDefault:"development"`
	Port   string `env:"PORT" envDefault:"8080"`

	DatabaseURL string `env:"DATABASE_URL,required,notEmpty"`
	RedisURL    string `env:"REDIS_URL,required,notEmpty"`

	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envDefault:""`
	TrustedProxyCIDRs  []string `env:"TRUSTED_PROXY_CIDRS" envDefault:""`

	RateLimitRequests int           `env:"RATE_LIMIT_REQUESTS" envDefault:"60"`
	RateLimitWindow   time.Duration `env:"RATE_LIMIT_WINDOW" envDefault:"1m"`

	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

	OTelEndpoint         string  `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:""`
	OTelServiceName      string  `env:"OTEL_SERVICE_NAME" envDefault:"sso-api"`
	OTelTraceSampleRatio float64 `env:"OTEL_TRACES_SAMPLER_ARG" envDefault:"1.0"`

	HTTPReadHeaderTimeout time.Duration `env:"HTTP_READ_HEADER_TIMEOUT" envDefault:"5s"`
	HTTPReadTimeout       time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"15s"`
	HTTPWriteTimeout      time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"30s"`
	HTTPIdleTimeout       time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`
	HTTPMaxBodyBytes      int64         `env:"HTTP_MAX_BODY_BYTES" envDefault:"1048576"`

	// TrustedProxies is derived from TrustedProxyCIDRs during Load.
	TrustedProxies []netip.Prefix `env:"-"`
}

// Load reads the environment into a Config and fails fast on a bad value.
func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("read environment: %w", err)
	}

	prefixes := make([]netip.Prefix, 0, len(cfg.TrustedProxyCIDRs))
	for _, cidr := range cfg.TrustedProxyCIDRs {
		prefix, err := netip.ParsePrefix(cidr)
		if err != nil {
			return Config{}, fmt.Errorf("TRUSTED_PROXY_CIDRS entry %q is not a CIDR: %w", cidr, err)
		}
		prefixes = append(prefixes, prefix)
	}
	cfg.TrustedProxies = prefixes

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) validate() error {
	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("PORT must be a number from 1 to 65535, got %q", c.Port)
	}
	if _, ok := parseLogLevel(c.LogLevel); !ok {
		return fmt.Errorf("LOG_LEVEL must be debug, info, warn, or error, got %q", c.LogLevel)
	}
	if c.RateLimitRequests < 1 {
		return fmt.Errorf("RATE_LIMIT_REQUESTS must be at least 1, got %d", c.RateLimitRequests)
	}
	if c.RateLimitWindow <= 0 {
		return fmt.Errorf("RATE_LIMIT_WINDOW must be positive, got %s", c.RateLimitWindow)
	}
	if c.OTelTraceSampleRatio < 0 || c.OTelTraceSampleRatio > 1 {
		return fmt.Errorf("OTEL_TRACES_SAMPLER_ARG must be between 0 and 1, got %v", c.OTelTraceSampleRatio)
	}
	if c.HTTPMaxBodyBytes < 1 {
		return fmt.Errorf("HTTP_MAX_BODY_BYTES must be positive, got %d", c.HTTPMaxBodyBytes)
	}
	for _, origin := range c.CORSAllowedOrigins {
		u, err := url.Parse(origin)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("CORS_ALLOWED_ORIGINS entry %q must be an absolute URL", origin)
		}
	}
	return nil
}

func parseLogLevel(value string) (slog.Level, bool) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(value)); err != nil {
		return 0, false
	}
	return level, true
}

// SlogLevel returns the parsed log level. Load has already validated it.
func (c Config) SlogLevel() slog.Level {
	level, _ := parseLogLevel(c.LogLevel)
	return level
}
