-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByGoogleID :one
SELECT * FROM users WHERE google_id = $1;

-- name: CreateUser :one
INSERT INTO users (id, email, password, name, role, google_id, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
RETURNING *;

-- name: UpdateUserRole :one
UPDATE users SET role = $2 WHERE id = $1 RETURNING *;

-- name: UpdateUserGoogleID :one
UPDATE users SET google_id = $2 WHERE id = $1 RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users SET password = $2 WHERE id = $1 RETURNING *;
