//go:build integration

package httphandler_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/ppabimanyu/sso/internal/auth"
	"github.com/ppabimanyu/sso/internal/httphandler"
	"github.com/ppabimanyu/sso/internal/middleware"
	"github.com/ppabimanyu/sso/internal/repository"
)

const (
	adminEmail    = "admin@example.com"
	adminPassword = "correct horse battery staple"
)

type appOptions struct {
	loginMaxFailures int
	loginWindow      time.Duration
	requestLimit     int
	idleTTL          time.Duration
	absoluteTTL      time.Duration
}

type testApp struct {
	server *httptest.Server
	client *http.Client
	pool   *pgxpool.Pool
	redis  *redis.Client
}

func newTestApp(t *testing.T, options appOptions) *testApp {
	t.Helper()

	if options.loginMaxFailures == 0 {
		options.loginMaxFailures = 10
	}
	if options.loginWindow == 0 {
		options.loginWindow = 15 * time.Minute
	}
	if options.idleTTL == 0 {
		options.idleTTL = time.Hour
	}
	if options.absoluteTTL == 0 {
		options.absoluteTTL = 2 * time.Hour
	}

	ctx := context.Background()
	container, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("app"),
		postgres.WithUsername("app"),
		postgres.WithPassword("app"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	connectionString, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	sqlDB, err := sql.Open("pgx", connectionString)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	require.NoError(t, goose.SetDialect("postgres"))
	goose.SetBaseFS(os.DirFS("../../db/migrations"))
	require.NoError(t, goose.Up(sqlDB, "."))

	pool, err := pgxpool.New(ctx, connectionString)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	redisContainer, err := testcontainers.Run(ctx, "redis:7-alpine",
		testcontainers.WithExposedPorts("6379/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("6379/tcp").WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = redisContainer.Terminate(ctx) })

	redisEndpoint, err := redisContainer.Endpoint(ctx, "")
	require.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisEndpoint})
	t.Cleanup(func() { _ = redisClient.Close() })

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	store := repository.NewPgAuth(pool)
	service, err := auth.NewService(
		auth.Config{SessionIdleTTL: options.idleTTL, SessionAbsoluteTTL: options.absoluteTTL},
		auth.Dependencies{
			Users:    store,
			Sessions: store,
			Audit:    store,
			Attempts: repository.NewLoginAttempts(redisClient, options.loginMaxFailures, options.loginWindow),
		},
		logger,
	)
	require.NoError(t, err)
	require.NoError(t, service.BootstrapAdmin(ctx, adminEmail, adminPassword))

	var rateLimit func(http.Handler) http.Handler
	if options.requestLimit > 0 {
		rateLimit = middleware.RateLimit(redisClient, options.requestLimit, time.Minute, logger)
	}

	router := httphandler.NewRouter(httphandler.RouterOptions{
		Logger:       logger,
		MaxBodyBytes: 1 << 20,
		RateLimit:    rateLimit,
		AuthService:  service,
	})
	server := httptest.NewTLSServer(router)
	t.Cleanup(server.Close)

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := server.Client()
	client.Jar = jar

	return &testApp{server: server, client: client, pool: pool, redis: redisClient}
}

func (a *testApp) closeRedis(t *testing.T) {
	t.Helper()
	require.NoError(t, a.redis.Close())
}

func (a *testApp) request(t *testing.T, method, path string, body any) *http.Response {
	t.Helper()

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(encoded)
	}

	request, err := http.NewRequest(method, a.server.URL+path, reader)
	require.NoError(t, err)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	request.Header.Set("Origin", a.server.URL)

	response, err := a.client.Do(request)
	require.NoError(t, err)
	t.Cleanup(func() { _ = response.Body.Close() })
	return response
}

func (a *testApp) login(t *testing.T, password string) *http.Response {
	t.Helper()
	return a.request(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    adminEmail,
		"password": password,
	})
}

func decodeBody[T any](t *testing.T, response *http.Response) T {
	t.Helper()

	var body T
	require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
	return body
}

func (a *testApp) auditCount(t *testing.T, action string, metadataKey, metadataValue string) int {
	t.Helper()

	// No Audit Event read API exists until #19, so the test inspects the table
	// directly. The login behavior itself is still asserted over HTTP.
	query := `SELECT count(*) FROM audit_events WHERE action = $1`
	arguments := []any{action}
	if metadataKey != "" {
		query += ` AND metadata->>$2 = $3`
		arguments = append(arguments, metadataKey, metadataValue)
	}

	var count int
	require.NoError(t, a.pool.QueryRow(context.Background(), query, arguments...).Scan(&count))
	return count
}

func (a *testApp) sessionCount(t *testing.T) int {
	t.Helper()

	var count int
	require.NoError(t, a.pool.QueryRow(context.Background(), `SELECT count(*) FROM sessions`).Scan(&count))
	return count
}

func TestBootstrapAdminCanLoginSeeProfileAndLogout(t *testing.T) {
	app := newTestApp(t, appOptions{})

	login := app.login(t, adminPassword)
	require.Equal(t, http.StatusOK, login.StatusCode)

	cookies := login.Cookies()
	require.Len(t, cookies, 1)
	cookie := cookies[0]
	assert.Equal(t, auth.SessionCookieName, cookie.Name)
	assert.NotEmpty(t, cookie.Value)
	assert.Empty(t, cookie.Domain)
	assert.Equal(t, "/", cookie.Path)
	assert.True(t, cookie.HttpOnly)
	assert.True(t, cookie.Secure)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	assert.Positive(t, cookie.MaxAge)

	loginBody := decodeBody[httphandler.LoginResponse](t, login)
	assert.Equal(t, adminEmail, loginBody.User.Email)
	assert.True(t, loginBody.User.IsAdmin)

	sessionResponse := app.request(t, http.MethodGet, "/api/v1/auth/session", nil)
	require.Equal(t, http.StatusOK, sessionResponse.StatusCode)
	sessionBody := decodeBody[httphandler.SessionResponse](t, sessionResponse)
	assert.True(t, sessionBody.Authenticated)
	require.NotNil(t, sessionBody.User)
	assert.True(t, sessionBody.User.IsAdmin)
	require.NotNil(t, sessionBody.ExpiresAt)

	me := app.request(t, http.MethodGet, "/api/v1/me", nil)
	require.Equal(t, http.StatusOK, me.StatusCode)
	profile := decodeBody[httphandler.UserResponse](t, me)
	assert.Equal(t, adminEmail, profile.Email)
	assert.True(t, profile.IsAdmin)
	assert.True(t, profile.EmailVerified)

	logout := app.request(t, http.MethodPost, "/api/v1/auth/logout", nil)
	require.Equal(t, http.StatusNoContent, logout.StatusCode)

	afterLogout := app.request(t, http.MethodGet, "/api/v1/me", nil)
	assert.Equal(t, http.StatusUnauthorized, afterLogout.StatusCode)

	sessionAfterLogout := app.request(t, http.MethodGet, "/api/v1/auth/session", nil)
	require.Equal(t, http.StatusOK, sessionAfterLogout.StatusCode)
	assert.False(t, decodeBody[httphandler.SessionResponse](t, sessionAfterLogout).Authenticated)

	assert.Equal(t, 1, app.auditCount(t, "auth.login.succeeded", "", ""))
	assert.Equal(t, 1, app.auditCount(t, "auth.logout.succeeded", "", ""))
	assert.Equal(t, 1, app.sessionCount(t))
}

func TestDisabledUserCannotLogin(t *testing.T) {
	app := newTestApp(t, appOptions{})

	// Admin user management does not exist until #19, so the disable is a setup
	// step. The assertion still goes through the login endpoint.
	_, err := app.pool.Exec(context.Background(), `UPDATE users SET disabled_at = now() WHERE email = $1`, adminEmail)
	require.NoError(t, err)

	response := app.login(t, adminPassword)
	require.Equal(t, http.StatusUnauthorized, response.StatusCode)
	assert.Empty(t, response.Cookies())
	assert.Zero(t, app.sessionCount(t))
	assert.Equal(t, 1, app.auditCount(t, "auth.login.failed", "reason", "user_disabled"))
}

func TestFailedLoginIsRecordedAndDoesNotRevealTheAccount(t *testing.T) {
	app := newTestApp(t, appOptions{})

	wrongPassword := app.login(t, "incorrect horse battery staple")
	require.Equal(t, http.StatusUnauthorized, wrongPassword.StatusCode)

	unknownEmail := app.request(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "missing@example.com",
		"password": "incorrect horse battery staple",
	})
	require.Equal(t, http.StatusUnauthorized, unknownEmail.StatusCode)

	wrongBody := decodeBody[map[string]any](t, wrongPassword)
	unknownBody := decodeBody[map[string]any](t, unknownEmail)
	assert.Equal(t, wrongBody["detail"], unknownBody["detail"])
	assert.Equal(t, "Email or password is incorrect.", wrongBody["detail"])

	assert.Equal(t, 2, app.auditCount(t, "auth.login.failed", "reason", "invalid_credentials"))
	assert.Zero(t, app.sessionCount(t))
}

func TestLoginLocksOutAfterRepeatedFailures(t *testing.T) {
	app := newTestApp(t, appOptions{})

	for range 10 {
		response := app.login(t, "incorrect horse battery staple")
		require.Equal(t, http.StatusUnauthorized, response.StatusCode)
	}

	blocked := app.login(t, adminPassword)
	require.Equal(t, http.StatusTooManyRequests, blocked.StatusCode)
	assert.NotEmpty(t, blocked.Header.Get("Retry-After"))
	assert.Zero(t, app.sessionCount(t))
	assert.Equal(t, 1, app.auditCount(t, "auth.login.failed", "reason", "rate_limited"))
}

func TestLoginEndpointIsRateLimited(t *testing.T) {
	app := newTestApp(t, appOptions{requestLimit: 2})

	for range 2 {
		response := app.login(t, "incorrect horse battery staple")
		require.Equal(t, http.StatusUnauthorized, response.StatusCode)
	}

	limited := app.login(t, "incorrect horse battery staple")
	require.Equal(t, http.StatusTooManyRequests, limited.StatusCode)
	assert.NotEmpty(t, limited.Header.Get("Retry-After"))
}

func TestLoginLockoutFallsBackToProcessMemoryWhenRedisIsDown(t *testing.T) {
	app := newTestApp(t, appOptions{loginMaxFailures: 3, loginWindow: 5 * time.Second})
	app.closeRedis(t)

	for range 2 {
		require.Equal(t, http.StatusUnauthorized, app.login(t, "incorrect horse battery staple").StatusCode)
	}
	// The first window is nearly over when the threshold is reached. The
	// fallback must start a full new window, like the Redis path does.
	time.Sleep(3 * time.Second)
	require.Equal(t, http.StatusUnauthorized, app.login(t, "incorrect horse battery staple").StatusCode)

	blocked := app.login(t, adminPassword)
	require.Equal(t, http.StatusTooManyRequests, blocked.StatusCode)
	retryAfter, err := strconv.Atoi(blocked.Header.Get("Retry-After"))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, retryAfter, 4)
}

func TestSuccessfulLoginResetsTheFailureCounter(t *testing.T) {
	app := newTestApp(t, appOptions{loginMaxFailures: 3})

	for range 2 {
		require.Equal(t, http.StatusUnauthorized, app.login(t, "incorrect horse battery staple").StatusCode)
	}
	require.Equal(t, http.StatusOK, app.login(t, adminPassword).StatusCode)

	for range 2 {
		require.Equal(t, http.StatusUnauthorized, app.login(t, "incorrect horse battery staple").StatusCode)
	}
	require.Equal(t, http.StatusOK, app.login(t, adminPassword).StatusCode)
}

func TestSessionIdleTTLExpiresAndCookieUsesAbsoluteTTL(t *testing.T) {
	app := newTestApp(t, appOptions{idleTTL: 500 * time.Millisecond, absoluteTTL: 2 * time.Hour})

	login := app.login(t, adminPassword)
	require.Equal(t, http.StatusOK, login.StatusCode)
	cookie := login.Cookies()[0]
	assert.Equal(t, int((2 * time.Hour).Seconds()), cookie.MaxAge)

	time.Sleep(700 * time.Millisecond)

	me := app.request(t, http.MethodGet, "/api/v1/me", nil)
	assert.Equal(t, http.StatusUnauthorized, me.StatusCode)
}
