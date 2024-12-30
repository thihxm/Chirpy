-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;


-- name: ResetChirps :exec
DELETE FROM chirps;


-- name: GetChirps :many
SELECT *
FROM chirps
WHERE (user_id = sqlc.narg('author_id') OR sqlc.narg('author_id') IS NULL)
ORDER BY CASE 
    WHEN sqlc.narg('sort') = 'desc' THEN created_at
END DESC, CASE 
    WHEN sqlc.narg('sort') = 'asc' OR sqlc.narg('sort') = '' THEN created_at
END ASC;


-- name: GetChirpByID :one
SELECT *
FROM chirps
WHERE id = $1;


-- name: DeleteChirpByID :exec
DELETE FROM chirps
WHERE id = $1 AND user_id = $2;
