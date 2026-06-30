-- name: GetAnswersByQuestionID :many
SELECT * FROM answers WHERE question_id = $1;

-- name: CreateAnswer :one
INSERT INTO answers (id, question_id, text, is_correct)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateAnswer :one
UPDATE answers SET text = $2, is_correct = $3 WHERE id = $1 RETURNING *;

-- name: DeleteAnswer :exec
DELETE FROM answers WHERE id = $1;

-- name: DeleteAnswersByQuestionID :exec
DELETE FROM answers WHERE question_id = $1;
