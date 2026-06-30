-- name: GetPlayerAnswerBySessionPlayerAndQuestion :one
SELECT * FROM player_answers
WHERE session_player_id = $1 AND question_id = $2;

-- name: GetPlayerAnswersBySessionPlayer :many
SELECT pa.*, q."order" AS question_order
FROM player_answers pa
JOIN questions q ON q.id = pa.question_id
WHERE pa.session_player_id = $1
ORDER BY q."order" ASC;

-- name: CreatePlayerAnswer :one
INSERT INTO player_answers (id, session_player_id, question_id, answered_at, is_correct, points)
VALUES ($1, $2, $3, NOW(), $4, $5)
RETURNING *;

-- name: ConnectPlayerAnswerToAnswer :exec
INSERT INTO player_answer_answers (player_answer_id, answer_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: GetVotesForQuestion :many
SELECT paa.answer_id, COUNT(*) AS votes
FROM player_answer_answers paa
JOIN player_answers pa ON pa.id = paa.player_answer_id
JOIN session_players sp ON sp.id = pa.session_player_id
WHERE pa.question_id = $1 AND sp.session_id = $2
GROUP BY paa.answer_id;

-- name: CountAnsweredPlayers :one
SELECT COUNT(*) FROM player_answers pa
JOIN session_players sp ON sp.id = pa.session_player_id
WHERE pa.question_id = $1 AND sp.session_id = $2;

-- name: GetSelectedAnswerIDs :many
SELECT answer_id FROM player_answer_answers WHERE player_answer_id = $1;
