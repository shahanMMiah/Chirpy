-- name: CreateChirps :one
INSERT INTO chirps(id, created_at, updated_at, body, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: ResetChirps :exec
DELETE FROM chirps;

-- name: GetAllChirps :many
SELECT * FROM chirps ORDER BY created_at ASC;

-- name: GetChirps :one
SELECT * FROM chirps WHERE id = $1 LIMIT 1;