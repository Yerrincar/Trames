-- name: SelectUserByUsername :one 
SELECT * FROM users WHERE username = ?;

-- name: SelectUserByID :one 
SELECT * FROM users WHERE id = ?;

-- name: InsertUser :one
INSERT INTO users (username, password_hash) VALUES (?,?) RETURNING *;
