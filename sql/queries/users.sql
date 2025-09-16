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

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUserFromEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET hashed_password = $4, email = $3, updated_at = $2
WHERE id = $1
RETURNING *;
