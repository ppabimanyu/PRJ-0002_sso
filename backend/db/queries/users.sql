-- name: GetUserByEmail :one
SELECT
    id,
    email,
    name,
    picture_url,
    is_admin,
    email_verified_at,
    disabled_at,
    created_at,
    updated_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT
    id,
    email,
    name,
    picture_url,
    is_admin,
    email_verified_at,
    disabled_at,
    created_at,
    updated_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateUser :exec
INSERT INTO users (id, email, name, picture_url, is_admin, email_verified_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetPasswordHash :one
SELECT password_hash
FROM password_credentials
WHERE user_id = $1;

-- name: UpsertPasswordHash :exec
INSERT INTO password_credentials (user_id, password_hash, updated_at)
VALUES ($1, $2, now())
ON CONFLICT (user_id) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    updated_at = now();
