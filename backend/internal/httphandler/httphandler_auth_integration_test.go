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
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
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
	"github.com/ppabimanyu/sso/internal/repository"
)

const (
	adminEmail    = "admin@example.com"
	adminPassword = "correct horse battery staple"
)

type appOptions struct {
	loginMaxFailures int
	idleTTL          time.Duration
	absoluteTTL      time.Duration
}

type testApp struct {
	server *httptest.Server
	client *http.Client
	pool   *pgxpool.Pool
}

func newTestApp(t *testing.T, options appOptions) *testApp {
	t.Helper()

	if options.loginMaxFailures == 0 {
		options.loginMaxFailures = 10
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

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	store := repository.NewPgAuth(pool)
	service, err := auth.NewService(
		auth.Config{SessionIdleTTL: options.idleTTL, SessionAbsoluteTTL: options.absoluteTTL},
		store,
		store,
		store,
		repository.NewLoginAttempts(redisClient, options.loginMaxFailures, 15*time.Minute),
		logger,
	)
	require.NoError(t, err)
	require.NoError(t, service.BootstrapAdmin(ctx, adminEmail, adminPassword))

	router := httphandler.NewRouter(httphandler.RouterOptions{
		Logger:       logger,
		MaxBodyBytes: 1 << 20,
		AuthService:  service,
	})
	server := httptest.NewTLSServer(router)
	t.Cleanup(server.Close)

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := server.Client()
	client.Jar = jar

	return &testApp{server: server, client: client, pool: pool}
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
