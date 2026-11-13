package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ppabimanyu/sso/internal/domain"
	"github.com/ppabimanyu/sso/internal/repository/sqlcgen"
)

// PgAuth implements the auth domain ports over PostgreSQL.
type PgAuth struct {
	pool    *pgxpool.Pool
	queries *sqlcgen.Queries
}

// NewPgAuth builds the auth repository.
func NewPgAuth(pool *pgxpool.Pool) *PgAuth {
	return &PgAuth{pool: pool, queries: sqlcgen.New(pool)}
}

// UserByEmail returns the active User with the given email.
func (r *PgAuth) UserByEmail(ctx context.Context, email string) (domain.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, mapPgError(err)
	}
	return userFromRow(row.ID, row.Email, row.Name, row.PictureUrl, row.IsAdmin, row.EmailVerifiedAt, row.DisabledAt, row.CreatedAt), nil
}

// UserByID returns the active User with the given id.
func (r *PgAuth) UserByID(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	row, err := r.queries.GetUserByID(ctx, pgUUID(userID))
	if err != nil {
		return domain.User{}, mapPgError(err)
	}
	return userFromRow(row.ID, row.Email, row.Name, row.PictureUrl, row.IsAdmin, row.EmailVerifiedAt, row.DisabledAt, row.CreatedAt), nil
}

// PasswordHash returns the stored Argon2id hash for a User.
func (r *PgAuth) PasswordHash(ctx context.Context, userID uuid.UUID) (string, error) {
	hash, err := r.queries.GetPasswordHash(ctx, pgUUID(userID))
	if err != nil {
		return "", mapPgError(err)
	}
	return hash, nil
}

// CreateUserWithPassword inserts the User and Credential atomically.
func (r *PgAuth) CreateUserWithPassword(ctx context.Context, user domain.User, passwordHash string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin create user: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queries := sqlcgen.New(tx)
	err = queries.CreateUser(ctx, sqlcgen.CreateUserParams{
		ID:              pgUUID(user.ID),
		Email:           user.Email,
		Name:            user.Name,
		PictureUrl:      pgText(user.PictureURL),
		IsAdmin:         user.IsAdmin,
		EmailVerifiedAt: pgTimePtr(user.EmailVerifiedAt),
	})
	if err != nil {
		return mapPgError(err)
	}
	err = queries.UpsertPasswordHash(ctx, sqlcgen.UpsertPasswordHashParams{
		UserID:       pgUUID(user.ID),
		PasswordHash: passwordHash,
	})
	if err != nil {
		return mapPgError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit create user: %w", err)
	}
	return nil
}

// CreateSessionWithAudit stores the Session and its login Audit Event in one
// transaction. A login cannot succeed without the audit record.
func (r *PgAuth) CreateSessionWithAudit(ctx context.Context, session domain.Session, tokenHash []byte, event domain.AuditEvent) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin create session: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queries := sqlcgen.New(tx)
	if err := r.createSession(ctx, queries, session, tokenHash); err != nil {
		return err
	}
	if err := insertAuditEvent(ctx, queries, event); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit create session: %w", err)
	}
	return nil
}

func (r *PgAuth) createSession(ctx context.Context, queries *sqlcgen.Queries, session domain.Session, tokenHash []byte) error {
	err := queries.CreateSession(ctx, sqlcgen.CreateSessionParams{
		ID:                pgUUID(session.ID),
		TokenHash:         tokenHash,
		UserID:            pgUUID(session.UserID),
		AuthMethod:        session.AuthMethod,
		Ip:                session.IP,
		UserAgent:         session.UserAgent,
		CreatedAt:         pgTime(session.CreatedAt),
		LastSeenAt:        pgTime(session.LastSeenAt),
		IdleExpiresAt:     pgTime(session.IdleExpiresAt),
		AbsoluteExpiresAt: pgTime(session.AbsoluteExpiresAt),
	})
	if err != nil {
		return mapPgError(err)
	}
	return nil
}

// SessionByTokenHash returns a live Session for the stored token hash.
func (r *PgAuth) SessionByTokenHash(ctx context.Context, tokenHash []byte) (domain.Session, error) {
	row, err := r.queries.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		return domain.Session{}, mapPgError(err)
	}
	return domain.Session{
		ID:                goUUID(row.ID),
		UserID:            goUUID(row.UserID),
		AuthMethod:        row.AuthMethod,
		IP:                row.Ip,
		UserAgent:         row.UserAgent,
		CreatedAt:         row.CreatedAt.Time,
		LastSeenAt:        row.LastSeenAt.Time,
		IdleExpiresAt:     row.IdleExpiresAt.Time,
		AbsoluteExpiresAt: row.AbsoluteExpiresAt.Time,
		RevokedAt:         goTimePtr(row.RevokedAt),
	}, nil
}

// TouchSession extends the idle expiry after activity.
func (r *PgAuth) TouchSession(ctx context.Context, sessionID uuid.UUID, lastSeenAt, idleExpiresAt time.Time) error {
	err := r.queries.TouchSession(ctx, sqlcgen.TouchSessionParams{
		ID:            pgUUID(sessionID),
		LastSeenAt:    pgTime(lastSeenAt),
		IdleExpiresAt: pgTime(idleExpiresAt),
	})
	if err != nil {
		return mapPgError(err)
	}
	return nil
}

// RevokeSession marks a Session revoked at the given time.
func (r *PgAuth) RevokeSession(ctx context.Context, sessionID uuid.UUID, revokedAt time.Time) error {
	err := r.queries.RevokeSession(ctx, sqlcgen.RevokeSessionParams{
		ID:        pgUUID(sessionID),
		RevokedAt: pgTime(revokedAt),
	})
	if err != nil {
		return mapPgError(err)
	}
	return nil
}

// RecordAuditEvent appends an Audit Event.
func (r *PgAuth) RecordAuditEvent(ctx context.Context, event domain.AuditEvent) error {
	return insertAuditEvent(ctx, r.queries, event)
}

func insertAuditEvent(ctx context.Context, queries *sqlcgen.Queries, event domain.AuditEvent) error {
	metadata := event.Metadata
	if metadata == nil {
		metadata = map[string]string{}
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("encode audit metadata: %w", err)
	}

	err = queries.InsertAuditEvent(ctx, sqlcgen.InsertAuditEventParams{
		ID:          pgUUID(event.ID),
		OccurredAt:  pgTime(event.OccurredAt),
		ActorType:   event.ActorType,
		ActorUserID: pgUUIDPtr(event.ActorUserID),
		Action:      event.Action,
		TargetType:  event.TargetType,
		TargetID:    pgUUIDPtr(event.TargetID),
		Ip:          event.IP,
		UserAgent:   event.UserAgent,
		RequestID:   event.RequestID,
		Metadata:    encoded,
	})
	if err != nil {
		return mapPgError(err)
	}
	return nil
}

func userFromRow(
	id pgtype.UUID,
	email, name string,
	pictureURL pgtype.Text,
	isAdmin bool,
	emailVerifiedAt, disabledAt, createdAt pgtype.Timestamptz,
) domain.User {
	return domain.User{
		ID:              goUUID(id),
		Email:           email,
		Name:            name,
		PictureURL:      pictureURL.String,
		IsAdmin:         isAdmin,
		EmailVerifiedAt: goTimePtr(emailVerifiedAt),
		DisabledAt:      goTimePtr(disabledAt),
		CreatedAt:       createdAt.Time,
	}
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func pgUUIDPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgUUID(*id)
}

func goUUID(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.Nil
	}
	return uuid.UUID(id.Bytes)
}

func pgTime(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: !value.IsZero()}
}

func pgTimePtr(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgTime(*value)
}

func goTimePtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	timeValue := value.Time
	return &timeValue
}

func pgText(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: value != ""}
}

func mapPgError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == "23505" {
		return domain.ErrConflict
	}
	return err
}
