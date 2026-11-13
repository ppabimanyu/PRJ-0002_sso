package domain

import (
	"time"

	"github.com/google/uuid"
)

// Session is an authenticated User's SSO Session. Only the hash of the opaque
// token is stored; the raw token lives in the browser cookie.
type Session struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	AuthMethod        string
	IP                string
	UserAgent         string
	CreatedAt         time.Time
	LastSeenAt        time.Time
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
	RevokedAt         *time.Time
}

// ActiveAt reports whether the Session may still authenticate a request.
func (s Session) ActiveAt(now time.Time) bool {
	return s.RevokedAt == nil && now.Before(s.IdleExpiresAt) && now.Before(s.AbsoluteExpiresAt)
}

// ExpiresAt returns the earlier of the idle and absolute expiry.
func (s Session) ExpiresAt() time.Time {
	if s.IdleExpiresAt.Before(s.AbsoluteExpiresAt) {
		return s.IdleExpiresAt
	}
	return s.AbsoluteExpiresAt
}
