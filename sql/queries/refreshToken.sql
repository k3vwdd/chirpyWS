-- name: GetUserFromRefreshToken :one
SELECT
    users.id,
    users.email,
    users.created_at,
    users.updated_at
FROM
    users
INNER JOIN
    refresh_tokens ON users.id = refresh_tokens.user_id
WHERE
    refresh_tokens.token = $1
    AND refresh_tokens.expires_at > NOW()
    AND (refresh_tokens.revoked_at IS NULL);
-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (token, user_id, expires_at, revoked_at)
VALUES ($1, $2, NOW() + INTERVAL '60 days', NULL);
-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE token = $1;

