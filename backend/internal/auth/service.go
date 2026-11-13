package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/ppabimanyu/sso/internal/domain"
)

// Config controls Session lifetime. Load validates the values so the service
// only has to defend against programmer error in tests.
type Config struct {
	SessionIdleTTL     time.Duration
	SessionAbsoluteTTL time.Duration
}

// RequestMeta is the transport data attached to an Audit Event.
type RequestMeta struct {
	IP        string
	UserAgent string
	RequestID string
}

// LoginInput is one password login attempt.
type LoginInput struct {
	Email    string
	Password string
	RequestMeta
}

// LoginResult is the authenticated User plus the raw token for the cookie.
type LoginResult struct {
	User    domain.User
	Session domain.Session
	Token   string
}

// BlockedError reports a temporary lockout. The handler turns it into 429 and
// a Retry-After header.
type BlockedError struct {
	RetryAfter time.Duration
}

func (e *BlockedError) Error() string {
	return fmt.Sprintf("login blocked for %s", e.RetryAfter.Round(time.Second))
}

// Dependencies are the domain ports the auth module needs. Naming them keeps
// the wiring readable when one adapter implements several ports.
type Dependencies struct {
	Users    domain.UserStore
	Sessions domain.SessionStore
	Audit    domain.AuditStore
	Attempts domain.LoginAttemptStore
}

// Service owns password login, Session validation, logout, and Admin
// bootstrap. It depends only on domain ports.
type Service struct {
	users       domain.UserStore
	sessions    domain.SessionStore
	audit       domain.AuditStore
	attempts    domain.LoginAttemptStore
	logger      *slog.Logger
	idleTTL     time.Duration
	absoluteTTL time.Duration
	now         func() time.Time
}

// NewService wires the authentication module. TTL validation fails fast so a
// misconfigured deployment never issues a Session with an impossible lifetime.
func NewService(cfg Config, deps Dependencies, logger *slog.Logger) (*Service, error) {
	if deps.Users == nil || deps.Sessions == nil || deps.Audit == nil || deps.Attempts == nil {
		return nil, fmt.Errorf("auth dependencies are incomplete")
	}
	if cfg.SessionIdleTTL <= 0 {
		return nil, fmt.Errorf("session idle TTL must be positive")
	}
	if cfg.SessionAbsoluteTTL <= 0 {
		return nil, fmt.Errorf("session absolute TTL must be positive")
	}
	if cfg.SessionIdleTTL > cfg.SessionAbsoluteTTL {
		return nil, fmt.Errorf("session idle TTL must not exceed the absolute TTL")
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &Service{
		users:       deps.Users,
		sessions:    deps.Sessions,
		audit:       deps.Audit,
		attempts:    deps.Attempts,
		logger:      logger,
		idleTTL:     cfg.SessionIdleTTL,
		absoluteTTL: cfg.SessionAbsoluteTTL,
		now:         time.Now,
	}, nil
}

// Login verifies a password, creates a Session, and records the attempt. The
// same sentinel is returned for an unknown email, a wrong password, and a
// disabled User so the response cannot enumerate accounts.
func (s *Service) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	email := NormalizeEmail(input.Email)

	if retryAfter, err := s.attempts.Blocked(ctx, email, input.IP); err != nil {
		s.logger.WarnContext(ctx, "login lockout unavailable", "error", err)
	} else if retryAfter > 0 {
		s.recordLoginAudit(ctx, nil, input.RequestMeta, domain.ActionLoginFailed, "rate_limited")
		return LoginResult{}, &BlockedError{RetryAfter: retryAfter}
	}

	user, userErr := s.users.UserByEmail(ctx, email)
	if userErr != nil && !errors.Is(userErr, domain.ErrNotFound) {
		return LoginResult{}, fmt.Errorf("find user: %w", userErr)
	}

	var passwordHash string
	if userErr == nil {
		hash, err := s.users.PasswordHash(ctx, user.ID)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return LoginResult{}, fmt.Errorf("read password credential: %w", err)
		}
		if err == nil {
			passwordHash = hash
		}
	}

	// Unknown email and missing credential both take the Argon2 path so the
	// response time does not reveal whether the account exists.
	if userErr != nil || passwordHash == "" {
		_, _ = VerifyPassword(input.Password, dummyPasswordHash())
		s.recordLoginFailure(ctx, nil, input, "invalid_credentials")
		return LoginResult{}, domain.ErrInvalidCredentials
	}

	valid, err := VerifyPassword(input.Password, passwordHash)
	if err != nil {
		return LoginResult{}, fmt.Errorf("verify password: %w", err)
	}
	if !valid {
		s.recordLoginFailure(ctx, &user, input, "invalid_credentials")
		return LoginResult{}, domain.ErrInvalidCredentials
	}
	if !user.Active() {
		s.recordLoginFailure(ctx, &user, input, "user_disabled")
		return LoginResult{}, domain.ErrInvalidCredentials
	}

	now := s.now()
	sessionID, err := uuid.NewV7()
	if err != nil {
		return LoginResult{}, fmt.Errorf("generate session id: %w", err)
	}
	token, tokenHash, err := NewSessionToken()
	if err != nil {
		return LoginResult{}, err
	}

	session := domain.Session{
		ID:                sessionID,
		UserID:            user.ID,
		AuthMethod:        domain.AuthMethodPassword,
		IP:                input.IP,
		UserAgent:         input.UserAgent,
		CreatedAt:         now,
		LastSeenAt:        now,
		IdleExpiresAt:     now.Add(s.idleTTL),
		AbsoluteExpiresAt: now.Add(s.absoluteTTL),
	}
	event, err := s.newAuditEvent(s.loginAuditEvent(&user, input.RequestMeta, domain.ActionLoginSucceeded, ""))
	if err != nil {
		return LoginResult{}, err
	}
	if err := s.sessions.CreateSessionWithAudit(ctx, session, tokenHash, event); err != nil {
		return LoginResult{}, fmt.Errorf("create session: %w", err)
	}

	if err := s.attempts.Reset(ctx, email, input.IP); err != nil {
		s.logger.WarnContext(ctx, "reset login lockout", "error", err)
	}

	return LoginResult{User: user, Session: session, Token: token}, nil
}

// AuthenticateSession resolves the cookie token to an active User and Session.
func (s *Service) AuthenticateSession(ctx context.Context, token string) (domain.User, domain.Session, error) {
	if token == "" {
		return domain.User{}, domain.Session{}, domain.ErrUnauthorized
	}

	session, err := s.sessions.SessionByTokenHash(ctx, HashSessionToken(token))
	if errors.Is(err, domain.ErrNotFound) {
		return domain.User{}, domain.Session{}, domain.ErrUnauthorized
	}
	if err != nil {
		return domain.User{}, domain.Session{}, fmt.Errorf("load session: %w", err)
	}

	now := s.now()
	if !session.ActiveAt(now) {
		if err := s.sessions.RevokeSession(ctx, session.ID, now); err != nil {
			s.logger.WarnContext(ctx, "revoke expired session", "error", err)
		}
		return domain.User{}, domain.Session{}, domain.ErrUnauthorized
	}

	user, err := s.users.UserByID(ctx, session.UserID)
	if errors.Is(err, domain.ErrNotFound) {
		_ = s.sessions.RevokeSession(ctx, session.ID, now)
		return domain.User{}, domain.Session{}, domain.ErrUnauthorized
	}
	if err != nil {
		return domain.User{}, domain.Session{}, fmt.Errorf("load session user: %w", err)
	}
	if !user.Active() {
		_ = s.sessions.RevokeSession(ctx, session.ID, now)
		return domain.User{}, domain.Session{}, domain.ErrUnauthorized
	}

	// Touching on every request would be a write per asset load. A minute of
	// granularity keeps idle expiry accurate enough without the churn.
	if now.Sub(session.LastSeenAt) >= time.Minute {
		idleExpiry := now.Add(s.idleTTL)
		if idleExpiry.After(session.AbsoluteExpiresAt) {
			idleExpiry = session.AbsoluteExpiresAt
		}
		if err := s.sessions.TouchSession(ctx, session.ID, now, idleExpiry); err != nil {
			s.logger.WarnContext(ctx, "touch session", "error", err)
		}
		session.LastSeenAt = now
		session.IdleExpiresAt = idleExpiry
	}

	return user, session, nil
}

// Logout revokes the Session behind a token. An unknown or already revoked
// token is a no-op so the endpoint is idempotent.
func (s *Service) Logout(ctx context.Context, token string, meta RequestMeta) error {
	if token == "" {
		return nil
	}

	session, err := s.sessions.SessionByTokenHash(ctx, HashSessionToken(token))
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load session for logout: %w", err)
	}

	now := s.now()
	if err := s.sessions.RevokeSession(ctx, session.ID, now); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	userID := session.UserID
	sessionID := session.ID
	s.recordAudit(ctx, domain.AuditEvent{
		ActorType:   domain.ActorTypeUser,
		ActorUserID: &userID,
		Action:      domain.ActionLogout,
		TargetType:  domain.TargetTypeSession,
		TargetID:    &sessionID,
		IP:          meta.IP,
		UserAgent:   meta.UserAgent,
		RequestID:   meta.RequestID,
		Metadata:    map[string]string{"method": session.AuthMethod},
	})
	return nil
}

// BootstrapAdmin creates the configured first Admin when that email is absent.
// It never changes an existing User, so a restart cannot reset a password.
func (s *Service) BootstrapAdmin(ctx context.Context, email, password string) error {
	email = NormalizeEmail(email)
	if email == "" {
		return nil
	}

	existing, err := s.users.UserByEmail(ctx, email)
	switch {
	case err == nil:
		if !existing.IsAdmin {
			s.logger.WarnContext(ctx, "bootstrap admin email belongs to a non-admin user", "user_id", existing.ID)
		}
		return nil
	case !errors.Is(err, domain.ErrNotFound):
		return fmt.Errorf("look up bootstrap admin: %w", err)
	}

	if password == "" {
		return fmt.Errorf("BOOTSTRAP_ADMIN_PASSWORD is required to create the bootstrap admin")
	}
	if err := ValidatePassword(password, email); err != nil {
		return fmt.Errorf("bootstrap admin password: %w", err)
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return err
	}

	now := s.now()
	userID, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate bootstrap admin id: %w", err)
	}
	user := domain.User{
		ID:              userID,
		Email:           email,
		IsAdmin:         true,
		EmailVerifiedAt: &now,
		CreatedAt:       now,
	}

	if err := s.users.CreateUserWithPassword(ctx, user, passwordHash); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			s.logger.InfoContext(ctx, "bootstrap admin already exists")
			return nil
		}
		return fmt.Errorf("create bootstrap admin: %w", err)
	}

	s.recordAudit(ctx, domain.AuditEvent{
		ActorType:  domain.ActorTypeSystem,
		Action:     domain.ActionAdminBootstrap,
		TargetType: domain.TargetTypeUser,
		TargetID:   &user.ID,
		Metadata:   map[string]string{"source": "configuration"},
	})
	return nil
}

func (s *Service) recordLoginFailure(ctx context.Context, user *domain.User, input LoginInput, reason string) {
	if err := s.attempts.RecordFailure(ctx, NormalizeEmail(input.Email), input.IP); err != nil {
		s.logger.WarnContext(ctx, "record login failure", "error", err)
	}

	s.recordLoginAudit(ctx, user, input.RequestMeta, domain.ActionLoginFailed, reason)
}

func (s *Service) recordLoginAudit(ctx context.Context, user *domain.User, meta RequestMeta, action, reason string) {
	s.recordAudit(ctx, s.loginAuditEvent(user, meta, action, reason))
}

func (s *Service) loginAuditEvent(user *domain.User, meta RequestMeta, action, reason string) domain.AuditEvent {
	event := domain.AuditEvent{
		ActorType: domain.ActorTypeAnonymous,
		Action:    action,
		IP:        meta.IP,
		UserAgent: meta.UserAgent,
		RequestID: meta.RequestID,
		Metadata:  map[string]string{"method": domain.AuthMethodPassword},
	}
	if user != nil {
		userID := user.ID
		event.ActorType = domain.ActorTypeUser
		event.ActorUserID = &userID
		event.TargetType = domain.TargetTypeUser
		event.TargetID = &userID
	}
	if action == domain.ActionLoginFailed {
		event.Metadata["reason"] = reason
	}
	return event
}

func (s *Service) recordAudit(ctx context.Context, event domain.AuditEvent) {
	prepared, err := s.newAuditEvent(event)
	if err != nil {
		s.logger.ErrorContext(ctx, "prepare audit event", "error", err)
		return
	}
	if err := s.audit.RecordAuditEvent(ctx, prepared); err != nil {
		s.logger.ErrorContext(ctx, "record audit event", "action", prepared.Action, "error", err)
	}
}

func (s *Service) newAuditEvent(event domain.AuditEvent) (domain.AuditEvent, error) {
	if event.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return domain.AuditEvent{}, fmt.Errorf("generate audit event id: %w", err)
		}
		event.ID = id
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = s.now()
	}
	return event, nil
}

// NormalizeEmail is the canonical form for lookup, rate-limit keys, and
// bootstrap configuration.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

var (
	dummyHashOnce sync.Once
	dummyHash     string
)

func dummyPasswordHash() string {
	dummyHashOnce.Do(func() {
		hash, err := HashPassword("timing-equalizer-not-a-real-password")
		if err != nil {
			return
		}
		dummyHash = hash
	})
	return dummyHash
}
