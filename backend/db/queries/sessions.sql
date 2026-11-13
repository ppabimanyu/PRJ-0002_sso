-- name: CreateSession :exec
INSERT INTO sessions (
    id,
    token_hash,
    user_id,
    auth_method,
    ip,
    user_agent,
    created_at,
    last_seen_at,
    idle_expires_at,
    absolute_expires_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetSessionByTokenHash :one
SELECT
    id,
    user_id,
    auth_method,
    ip,
    user_agent,
    created_at,
    last_seen_at,
    idle_expires_at,
    absolute_expires_at,
    revoked_at
FROM sessions
WHERE token_hash = $1 AND revoked_at IS NULL;

-- name: TouchSession :exec
UPDATE sessions
SET last_seen_at = $2,
    idle_expires_at = $3
WHERE id = $1 AND revoked_at IS NULL;

-- name: RevokeSession :exec
UPDATE sessions
SET revoked_at = $2
WHERE id = $1 AND revoked_at IS NULL;
