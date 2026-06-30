-- name: GetSessionByID :one
SELECT * FROM quiz_sessions WHERE id = $1;

-- name: GetSessionByRoomCode :one
SELECT * FROM quiz_sessions WHERE room_code = $1;

-- name: GetLatestSessionByQuizID :one
SELECT * FROM quiz_sessions
WHERE quiz_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateSession :one
INSERT INTO quiz_sessions (id, quiz_id, room_code, status, created_at)
VALUES ($1, $2, $3, 'WAITING', NOW())
RETURNING *;

-- name: UpdateSessionStatus :one
UPDATE quiz_sessions SET status = $2 WHERE id = $1 RETURNING *;

-- name: StartSession :one
UPDATE quiz_sessions SET status = 'ACTIVE', started_at = NOW() WHERE id = $1 RETURNING *;

-- name: FinishSession :exec
UPDATE quiz_sessions SET status = 'FINISHED' WHERE id = $1;

-- name: DeleteSession :exec
DELETE FROM quiz_sessions WHERE id = $1;

-- name: GetFinishedSessionsByQuizID :many
SELECT id, quiz_id, room_code, status, started_at, created_at FROM quiz_sessions
WHERE quiz_id = $1 AND status = 'FINISHED' AND started_at IS NOT NULL
ORDER BY started_at DESC;
