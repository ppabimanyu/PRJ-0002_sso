package domain

import "context"

// HealthChecker reports whether a dependency can serve requests.
// Repositories implement it for PostgreSQL, Redis, or any later dependency.
type HealthChecker interface {
	Ping(ctx context.Context) error
}
