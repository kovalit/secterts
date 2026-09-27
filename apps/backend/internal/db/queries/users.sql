-- Example sqlc queries (not used at runtime; see sqlc.yaml note).

-- name: GetUserByEmail :one
SELECT id, email, password_hash, email_verified, email_2fa_enabled, is_active, created_at, updated_at
FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id, email, password_hash, email_verified, email_2fa_enabled, is_active, created_at, updated_at;

-- name: GetSessionByTokenHash :one
SELECT id, user_id, refresh_token_hash, user_agent, ip, expires_at, revoked_at, created_at
FROM sessions
WHERE refresh_token_hash = $1 AND revoked_at IS NULL AND expires_at > now();
