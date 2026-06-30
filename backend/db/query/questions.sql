-- name: GetQuestionByID :one
SELECT * FROM questions WHERE id = $1;

-- name: GetQuestionsByQuizID :many
SELECT * FROM questions WHERE quiz_id = $1 ORDER BY "order" ASC;

-- name: CreateQuestion :one
INSERT INTO questions (id, quiz_id, text, image_url, type, "order", tags, time_limit, points)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateQuestion :one
UPDATE questions SET
    text       = $2,
    image_url  = $3,
    type       = $4,
    "order"    = $5,
    tags       = $6,
    time_limit = $7,
    points     = $8
WHERE id = $1
RETURNING *;

-- name: DeleteQuestion :exec
DELETE FROM questions WHERE id = $1;
