//go:build integration

package repository_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/ppabimanyu/sso/internal/repository"
)

func TestPgHealthRunsGeneratedQueryAgainstPostgres(t *testing.T) {
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

	db, err := sql.Open("pgx", connectionString)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	require.NoError(t, goose.SetDialect("postgres"))
	goose.SetBaseFS(os.DirFS("../../db/migrations"))
	require.NoError(t, goose.Up(db, "."))

	pool, err := pgxpool.New(ctx, connectionString)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	health := repository.NewPgHealth(pool)
	require.NoError(t, health.Ping(ctx))
}
