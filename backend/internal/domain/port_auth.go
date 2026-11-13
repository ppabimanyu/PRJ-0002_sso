package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UserStore reads and writes Users and their password Credentials.
type UserStore interface {
	UserByEmail(ctx context.Context, email string) (User, error)
	UserByID(ctx context.Context, userID uuid.UUID) (User, error)
	PasswordHash(ctx context.Context, userID uuid.UUID) (string, error)
	CreateUserWithPassword(ctx context.Context, user User, passwordHash string) error
}

// SessionStore persists hashed Session tokens.
type SessionStore interface {
	CreateSession(ctx context.Context, session Session, tokenHash []byte) error
	SessionByTokenHash(ctx context.Context, tokenHash []byte) (Session, error)
	TouchSession(ctx context.Context, sessionID uuid.UUID, lastSeenAt, idleExpiresAt time.Time) error
	RevokeSession(ctx context.Context, sessionID uuid.UUID, revokedAt time.Time) error
}

// AuditStore appends Audit Events.
type AuditStore interface {
	RecordAuditEvent(ctx context.Context, event AuditEvent) error
}

// LoginAttemptStore bounds password guessing per account and client IP. It
// returns the remaining lockout duration when the pair is blocked.
type LoginAttemptStore interface {
	Blocked(ctx context.Context, email, ip string) (time.Duration, error)
	RecordFailure(ctx context.Context, email, ip string) error
	Reset(ctx context.Context, email, ip string) error
}
