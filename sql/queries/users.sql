-- name: CreateUser :one
INSERT INTO users(hashed_password, email)
VALUES (
    $1,
    $2
)
RETURNING *;

-- name: ResetUser :exec
TRUNCATE users CASCADE;

-- name: UserLogin :one
SELECT *
FROM users
WHERE email = $1;

-- name: UpdateUser :exec
UPDATE users
SET
    hashed_password = $1,
    email = $2
WHERE id = $3;

-- name: UpgradeUser :exec
UPDATE users
SET
    is_chirpy_red = true
WHERE id = $1;