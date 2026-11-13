package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ppabimanyu/sso/internal/domain"
)

type fakeUsers struct {
	user         domain.User
	passwordHash string
	found        bool
	createErr    error
	created      []domain.User
	createdHash  string
}

func (f *fakeUsers) UserByEmail(context.Context, string) (domain.User, error) {
	if !f.found {
		return domain.User{}, domain.ErrNotFound
	}
	return f.user, nil
}

func (f *fakeUsers) UserByID(context.Context, uuid.UUID) (domain.User, error) {
	if !f.found {
		return domain.User{}, domain.ErrNotFound
	}
	return f.user, nil
}

func (f *fakeUsers) PasswordHash(context.Context, uuid.UUID) (string, error) {
	if f.passwordHash == "" {
		return "", domain.ErrNotFound
	}
	return f.passwordHash, nil
}

func (f *fakeUsers) CreateUserWithPassword(_ context.Context, user domain.User, hash string) error {
	if f.createErr != nil {
		return f.createErr
	}
	if f.found {
		return domain.ErrConflict
	}
	f.found = true
	f.user = user
	f.created = append(f.created, user)
	f.createdHash = hash
	return nil
}

type fakeSessions struct {
	sessions map[string]domain.Session
	revoked  map[uuid.UUID]bool
	touched  map[uuid.UUID]bool
	events   []domain.AuditEvent
}

func newFakeSessions() *fakeSessions {
	return &fakeSessions{
		sessions: map[string]domain.Session{},
		revoked:  map[uuid.UUID]bool{},
		touched:  map[uuid.UUID]bool{},
	}
}

func (f *fakeSessions) createSession(_ context.Context, session domain.Session, tokenHash []byte) error {
	f.sessions[string(tokenHash)] = session
	return nil
}

func (f *fakeSessions) CreateSessionWithAudit(ctx context.Context, session domain.Session, tokenHash []byte, event domain.AuditEvent) error {
	if err := f.createSession(ctx, session, tokenHash); err != nil {
		return err
	}
	f.events = append(f.events, event)
	return nil
}

func (f *fakeSessions) SessionByTokenHash(_ context.Context, tokenHash []byte) (domain.Session, error) {
	session, ok := f.sessions[string(tokenHash)]
	if !ok || f.revoked[session.ID] {
		return domain.Session{}, domain.ErrNotFound
	}
	return session, nil
}

func (f *fakeSessions) TouchSession(_ context.Context, id uuid.UUID, lastSeenAt, idleExpiresAt time.Time) error {
	session := f.sessions[sessionKey(f.sessions, id)]
	session.LastSeenAt = lastSeenAt
	session.IdleExpiresAt = idleExpiresAt
	f.sessions[sessionKey(f.sessions, id)] = session
	f.touched[id] = true
	return nil
}

func (f *fakeSessions) RevokeSession(_ context.Context, id uuid.UUID, _ time.Time) error {
	f.revoked[id] = true
	return nil
}

func sessionKey(sessions map[string]domain.Session, id uuid.UUID) string {
	for key, session := range sessions {
		if session.ID == id {
			return key
		}
	}
	return ""
}

type fakeAudit struct {
	events []domain.AuditEvent
}

func (f *fakeAudit) RecordAuditEvent(_ context.Context, event domain.AuditEvent) error {
	f.events = append(f.events, event)
	return nil
}

func (f *fakeAudit) lastAction() string {
	if len(f.events) == 0 {
		return ""
	}
	return f.events[len(f.events)-1].Action
}

type fakeAttempts struct {
	blockedFor time.Duration
	blockedErr error
	failures   int
	resets     int
}

func (f *fakeAttempts) Blocked(context.Context, string, string) (time.Duration, error) {
	return f.blockedFor, f.blockedErr
}

func (f *fakeAttempts) RecordFailure(context.Context, string, string) error {
	f.failures++
	return nil
}

func (f *fakeAttempts) Reset(context.Context, string, string) error {
	f.resets++
	return nil
}

type fakeClock struct {
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	return c.now
}

func newTestService(t *testing.T, users *fakeUsers, sessions *fakeSessions, audit *fakeAudit, attempts *fakeAttempts, clock *fakeClock) *Service {
	t.Helper()

	service, err := NewService(
		Config{SessionIdleTTL: time.Hour, SessionAbsoluteTTL: 24 * time.Hour},
		Dependencies{Users: users, Sessions: sessions, Audit: audit, Attempts: attempts},
		slog.New(slog.NewTextHandler(testWriter{t}, nil)),
	)
	require.NoError(t, err)
	service.now = clock.Now
	return service
}

type testWriter struct {
	t *testing.T
}

func (w testWriter) Write(buffer []byte) (int, error) {
	w.t.Log(string(buffer))
	return len(buffer), nil
}

func activeUser(t *testing.T) (domain.User, string) {
	t.Helper()

	hash, err := HashPassword("correct horse battery staple")
	require.NoError(t, err)
	return domain.User{
		ID:        uuid.Must(uuid.NewV7()),
		Email:     "admin@example.com",
		Name:      "Admin",
		IsAdmin:   true,
		CreatedAt: time.Now(),
	}, hash
}

func TestLoginCreatesSessionAndAudit(t *testing.T) {
	ctx := context.Background()
	user, hash := activeUser(t)
	users := &fakeUsers{user: user, passwordHash: hash, found: true}
	sessions := newFakeSessions()
	audit := &fakeAudit{}
	attempts := &fakeAttempts{}
	service := newTestService(t, users, sessions, audit, attempts, &fakeClock{now: time.Now()})

	result, err := service.Login(ctx, LoginInput{
		Email:       "Admin@Example.com",
		Password:    "correct horse battery staple",
		RequestMeta: RequestMeta{IP: "203.0.113.10", UserAgent: "test", RequestID: "request-1"},
	})
	require.NoError(t, err)

	assert.Equal(t, hash, users.passwordHash)
	assert.NotEmpty(t, result.Token)
	assert.Equal(t, user.ID, result.Session.UserID)
	require.Len(t, sessions.sessions, 1)
	require.Len(t, sessions.events, 1)
	assert.Equal(t, domain.ActionLoginSucceeded, sessions.events[0].Action)
	assert.Equal(t, 1, attempts.resets)
	assert.Equal(t, "203.0.113.10", sessions.events[0].IP)
}

func TestLoginRejectsUnknownEmailAndRecordsFailure(t *testing.T) {
	ctx := context.Background()
	users := &fakeUsers{}
	audit := &fakeAudit{}
	attempts := &fakeAttempts{}
	service := newTestService(t, users, newFakeSessions(), audit, attempts, &fakeClock{now: time.Now()})

	_, err := service.Login(ctx, LoginInput{Email: "missing@example.com", Password: "not-the-password"})
	require.ErrorIs(t, err, domain.ErrInvalidCredentials)

	assert.Equal(t, 1, attempts.failures)
	assert.Equal(t, domain.ActionLoginFailed, audit.lastAction())
	assert.Equal(t, "invalid_credentials", audit.events[0].Metadata["reason"])
}

func TestLoginRejectsDisabledUser(t *testing.T) {
	ctx := context.Background()
	user, hash := activeUser(t)
	disabledAt := time.Now()
	user.DisabledAt = &disabledAt
	users := &fakeUsers{user: user, passwordHash: hash, found: true}
	audit := &fakeAudit{}
	attempts := &fakeAttempts{}
	sessions := newFakeSessions()
	service := newTestService(t, users, sessions, audit, attempts, &fakeClock{now: time.Now()})

	_, err := service.Login(ctx, LoginInput{Email: user.Email, Password: "correct horse battery staple"})
	require.ErrorIs(t, err, domain.ErrInvalidCredentials)

	assert.Equal(t, 1, attempts.failures)
	assert.Empty(t, sessions.sessions)
	assert.Equal(t, "user_disabled", audit.events[0].Metadata["reason"])
}

func TestLoginReturnsBlockedError(t *testing.T) {
	ctx := context.Background()
	users := &fakeUsers{}
	audit := &fakeAudit{}
	attempts := &fakeAttempts{blockedFor: 15 * time.Minute}
	service := newTestService(t, users, newFakeSessions(), audit, attempts, &fakeClock{now: time.Now()})

	_, err := service.Login(ctx, LoginInput{Email: "person@example.com", Password: "whatever"})
	var blocked *BlockedError
	require.ErrorAs(t, err, &blocked)
	assert.Equal(t, 15*time.Minute, blocked.RetryAfter)
	assert.Equal(t, "rate_limited", audit.events[0].Metadata["reason"])
	assert.Equal(t, 0, attempts.failures)
}

func TestLogoutRevokesSessionAndAudits(t *testing.T) {
	ctx := context.Background()
	user, hash := activeUser(t)
	sessions := newFakeSessions()
	audit := &fakeAudit{}
	service := newTestService(t, &fakeUsers{user: user, passwordHash: hash, found: true}, sessions, audit, &fakeAttempts{}, &fakeClock{now: time.Now()})

	result, err := service.Login(ctx, LoginInput{Email: user.Email, Password: "correct horse battery staple"})
	require.NoError(t, err)

	err = service.Logout(ctx, result.Token, RequestMeta{IP: "203.0.113.10"})
	require.NoError(t, err)

	assert.True(t, sessions.revoked[result.Session.ID])
	require.Len(t, sessions.events, 1)
	assert.Equal(t, domain.ActionLogout, audit.lastAction())

	_, _, err = service.AuthenticateSession(ctx, result.Token)
	require.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestAuthenticateSessionRejectsExpiredSession(t *testing.T) {
	ctx := context.Background()
	user, _ := activeUser(t)
	sessions := newFakeSessions()
	audit := &fakeAudit{}
	clock := &fakeClock{now: time.Now()}
	service := newTestService(t, &fakeUsers{user: user, found: true}, sessions, audit, &fakeAttempts{}, clock)

	token, hash, err := NewSessionToken()
	require.NoError(t, err)
	expired := domain.Session{
		ID:                uuid.Must(uuid.NewV7()),
		UserID:            user.ID,
		AuthMethod:        "password",
		IdleExpiresAt:     clock.now.Add(-time.Minute),
		AbsoluteExpiresAt: clock.now.Add(time.Hour),
	}
	require.NoError(t, sessions.createSession(ctx, expired, hash))

	_, _, err = service.AuthenticateSession(ctx, token)
	require.ErrorIs(t, err, domain.ErrUnauthorized)
	assert.True(t, sessions.revoked[expired.ID])
}

func TestBootstrapAdminCreatesUserThenSkipsExisting(t *testing.T) {
	ctx := context.Background()
	users := &fakeUsers{}
	service := newTestService(t, users, newFakeSessions(), &fakeAudit{}, &fakeAttempts{}, &fakeClock{now: time.Now()})

	err := service.BootstrapAdmin(ctx, "Admin@Example.com", "correct horse battery staple")
	require.NoError(t, err)
	require.Len(t, users.created, 1)
	assert.Equal(t, "admin@example.com", users.created[0].Email)
	assert.True(t, users.created[0].IsAdmin)
	assert.NotNil(t, users.created[0].EmailVerifiedAt)
	assert.NotEmpty(t, users.createdHash)

	err = service.BootstrapAdmin(ctx, "admin@example.com", "a-different-password")
	require.NoError(t, err)
	assert.Len(t, users.created, 1)
}

func TestBootstrapAdminRequiresPasswordToCreate(t *testing.T) {
	users := &fakeUsers{}
	service := newTestService(t, users, newFakeSessions(), &fakeAudit{}, &fakeAttempts{}, &fakeClock{now: time.Now()})

	err := service.BootstrapAdmin(context.Background(), "admin@example.com", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "BOOTSTRAP_ADMIN_PASSWORD")
	assert.Empty(t, users.created)
}

func TestSessionCookieIsHostOnlyAndSecure(t *testing.T) {
	now := time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)

	cookie := SessionCookie("token", now.Add(90*24*time.Hour), now)

	assert.Equal(t, SessionCookieName, cookie.Name)
	assert.Equal(t, "token", cookie.Value)
	assert.Empty(t, cookie.Domain)
	assert.Equal(t, "/", cookie.Path)
	assert.True(t, cookie.HttpOnly)
	assert.True(t, cookie.Secure)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	assert.Equal(t, 90*24*60*60, cookie.MaxAge)
}

func TestNewSessionTokenIsOpaqueAndHashed(t *testing.T) {
	first, firstHash, err := NewSessionToken()
	require.NoError(t, err)
	second, secondHash, err := NewSessionToken()
	require.NoError(t, err)

	assert.NotEqual(t, first, second)
	assert.NotEqual(t, firstHash, secondHash)
	assert.Equal(t, firstHash, HashSessionToken(first))
	assert.Len(t, firstHash, 32)
}

func TestNewServiceRejectsIdleTTLLongerThanAbsolute(t *testing.T) {
	_, err := NewService(
		Config{SessionIdleTTL: 48 * time.Hour, SessionAbsoluteTTL: 24 * time.Hour},
		Dependencies{Users: &fakeUsers{}, Sessions: newFakeSessions(), Audit: &fakeAudit{}, Attempts: &fakeAttempts{}},
		nil,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "idle TTL")
}

func TestLoginTreatsMalformedStoredHashAsInternalError(t *testing.T) {
	ctx := context.Background()
	user, _ := activeUser(t)
	users := &fakeUsers{user: user, passwordHash: "broken", found: true}
	service := newTestService(t, users, newFakeSessions(), &fakeAudit{}, &fakeAttempts{}, &fakeClock{now: time.Now()})

	_, err := service.Login(ctx, LoginInput{Email: user.Email, Password: "correct horse battery staple"})
	require.Error(t, err)
	assert.False(t, errors.Is(err, domain.ErrInvalidCredentials))
}
