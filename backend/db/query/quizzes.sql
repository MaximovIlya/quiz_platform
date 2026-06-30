-- name: GetQuizByID :one
SELECT * FROM quizzes WHERE id = $1;

-- name: ListQuizzesByAuthor :many
SELECT * FROM quizzes
WHERE author_id = $1 AND archived = FALSE
ORDER BY created_at DESC;

-- name: CreateQuiz :one
INSERT INTO quizzes (id, title, description, category, time_per_question, points_per_question, scoring, difficulty, tags, cover_image_url, archived, author_id, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, FALSE, $11, NOW())
RETURNING *;

-- name: UpdateQuiz :one
UPDATE quizzes SET
    title               = $2,
    description         = $3,
    category            = $4,
    time_per_question   = $5,
    points_per_question = $6,
    scoring             = $7,
    difficulty          = $8,
    tags                = $9,
    cover_image_url     = $10
WHERE id = $1
RETURNING *;

-- name: ArchiveQuiz :exec
UPDATE quizzes SET archived = TRUE WHERE id = $1;

-- name: DeleteQuiz :exec
DELETE FROM quizzes WHERE id = $1;
