package domain

import (
	"time"

	"github.com/google/uuid"
)

// Actor types in an Audit Event. An anonymous actor is an unauthenticated
// request whose identity is unknown or deliberately not recorded.
const (
	ActorTypeUser      = "user"
	ActorTypeAnonymous = "anonymous"
	ActorTypeSystem    = "system"
)

// Audit actions recorded by the authentication surface.
const (
	ActionLoginSucceeded = "auth.login.succeeded"
	ActionLoginFailed    = "auth.login.failed"
	ActionLogout         = "auth.logout.succeeded"
	ActionAdminBootstrap = "auth.admin.bootstrap"
)

// Target types referenced by an Audit Event.
const (
	TargetTypeUser    = "user"
	TargetTypeSession = "session"
)

// AuditEvent is an append-only record of an action that touched identity or
// access.
type AuditEvent struct {
	ID          uuid.UUID
	OccurredAt  time.Time
	ActorType   string
	ActorUserID *uuid.UUID
	Action      string
	TargetType  string
	TargetID    *uuid.UUID
	IP          string
	UserAgent   string
	RequestID   string
	Metadata    map[string]string
}
