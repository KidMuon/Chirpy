-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (gen_random_uuid(), $1, $2, $3, $4)
RETURNING *;

-- name: GetAllChirps :many
SELECT * FROM chirps
ORDER BY created_at ASC;

-- name: GetAllChirpsByAuthor :many
SELECT * FROM chirps
WHERE user_id = $1
ORDER BY created_at ASC;

-- name: GetChirpByID :one
SELECT * FROM chirps
WHERE id = $1;

-- name: DeleteChirpByID :one
DELETE FROM chirps
WHERE id = $1
RETURNING *;
