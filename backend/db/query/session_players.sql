-- name: GetSessionPlayerByID :one
SELECT * FROM session_players WHERE id = $1;

-- name: GetSessionPlayerByUserAndSession :one
SELECT * FROM session_players WHERE session_id = $1 AND user_id = $2;

-- name: GetSessionPlayers :many
SELECT * FROM session_players WHERE session_id = $1;

-- name: GetLeaderboard :many
SELECT
    sp.id   AS session_player_id,
    sp.score,
    u.id    AS user_id,
    u.name
FROM session_players sp
JOIN users u ON u.id = sp.user_id
WHERE sp.session_id = $1
ORDER BY sp.score DESC;

-- name: CreateSessionPlayer :one
INSERT INTO session_players (id, session_id, user_id, score, joined_at)
VALUES ($1, $2, $3, 0, NOW())
RETURNING *;

-- name: IncrementPlayerScore :one
UPDATE session_players SET score = score + $2 WHERE id = $1 RETURNING *;
