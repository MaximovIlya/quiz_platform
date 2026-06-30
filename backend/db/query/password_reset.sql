-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (id, email, token, expires_at, created_at)
VALUES ($1, $2, $3, $4, NOW())
RETURNING *;

-- name: GetPasswordResetToken :one
SELECT * FROM password_reset_tokens WHERE token = $1;

-- name: DeletePasswordResetToken :exec
DELETE FROM password_reset_tokens WHERE token = $1;

-- name: DeletePasswordResetTokensByEmail :exec
DELETE FROM password_reset_tokens WHERE email = $1;
