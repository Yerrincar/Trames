-- name: SelectProjectsByUserAndProject :many
SELECT * FROM projects WHERE user_id = ?; 

-- name: InsertProjectsByUserAndProject :one
INSERT INTO projects (project, description, status) VALUES (?, ?, ?) RETURNING *;

-- name: UpdateProjectsByUserAndProject :exec
UPDATE projects SET project = ?, description = ?, status = ? WHERE id = ? AND user_id = ?;

-- name: DeleteProjects :many
DELETE FROM projects WHERE id IN (sqlc.slice('ids')) RETURNING *;
