-- name: InsertAuditEvent :exec
INSERT INTO audit_events (
    id,
    occurred_at,
    actor_type,
    actor_user_id,
    action,
    target_type,
    target_id,
    ip,
    user_agent,
    request_id,
    metadata
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
