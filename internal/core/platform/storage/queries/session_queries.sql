-- name: CreateSession :one
INSERT INTO sessions (user_id, token_hash, expires_at) VALUES (?, ?, ?) RETURNING *;

-- name: FindSessionByTokenHash :one
SELECT * FROM sessions WHERE token_hash = ? AND expires_at > NOW();

-- name: DeleteSessionByTokenHash :exec
DELETE FROM sessions WHERE token_hash = ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at < NOW();

-- name: DeleteSessionsByUserID :exec
DELETE FROM sessions WHERE user_id = ?;
