-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;
-- name: DeleteAllUsers :exec
DELETE FROM users;
-- name: GetUserByEmail :one
SELECT id, created_at, updated_at, email, hashed_password
FROM users
WHERE email = $1
LIMIT 1;
-- name: UpdateUserEmailAndPassword :exec
UPDATE users
SET
    email = $1,
    hashed_password = $2,
    updated_at = NOW()
WHERE
    id = $3;
-- name: GetUserByID :one
SELECT
    id,
    created_at,
    updated_at,
    email,
    hashed_password
FROM
    users
WHERE
    id = $1
LIMIT 1;

