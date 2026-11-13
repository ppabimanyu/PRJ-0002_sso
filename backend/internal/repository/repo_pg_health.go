// Package repository implements domain ports over PostgreSQL and external
// services. It holds no business rules.
package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ppabimanyu/sso/internal/repository/sqlcgen"
)

// PgHealth implements domain.HealthChecker through the generated sqlc query.
// It proves the sqlc pipeline without adding a domain table.
type PgHealth struct {
	queries *sqlcgen.Queries
}

// NewPgHealth builds the PostgreSQL health checker.
func NewPgHealth(pool *pgxpool.Pool) *PgHealth {
	return &PgHealth{queries: sqlcgen.New(pool)}
}

// Ping satisfies domain.HealthChecker through the generated sqlc query.
func (h *PgHealth) Ping(ctx context.Context) error {
	if _, err := h.queries.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}
	return nil
}
