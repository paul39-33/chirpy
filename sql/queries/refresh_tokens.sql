-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, user_id, updated_at, expires_at)
VALUES (
    $1,
    $2,
    $3,
    $4
) RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT *
FROM refresh_tokens
WHERE token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET
    updated_at = now(),
    revoked_at = now()
WHERE token = $1;