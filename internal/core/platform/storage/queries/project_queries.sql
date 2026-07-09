-- name: SelectProjectsByUserAndProject :many
SELECT * FROM projects WHERE user_id = ?; 

-- name: SelectProjectsByUserAndProjectId :one
SELECT * FROM projects WHERE user_id = ? AND id = ?; 

-- name: InsertProjectsByUserAndProject :one
INSERT INTO projects (user_id, project, description, status) VALUES (?, ?, ?, ?) RETURNING *;

-- name: UpdateProjectsByUserAndProject :exec
UPDATE projects SET project = ?, description = ?, status = ? WHERE id = ? AND user_id = ?;

-- name: SelectIdByProject :one
SELECT id FROM projects WHERE user_id = ? AND project = ?;

-- name: DeleteProjects :one
DELETE FROM projects WHERE id = ? AND user_id = ? RETURNING *;
