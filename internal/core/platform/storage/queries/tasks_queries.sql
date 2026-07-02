-- name: SelectTasksByUserAndProject :many
SELECT * FROM tasks WHERE user_id = ? AND project_id = ?; 

-- name: InsertTasksByUserAndProject :one
INSERT INTO tasks (user_id, project_id, task, description, status, priority) VALUES (?, ?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateTasksByUserAndProject :exec
UPDATE tasks SET task = ?, description = ?, status = ?, priority = ? WHERE id = ? AND user_id = ?;

-- name: DeleteTask :many
DELETE FROM tasks WHERE id IN (sqlc.slice('ids')) RETURNING *;
