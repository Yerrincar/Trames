-- name: SelectSubProjectsByUserAndProject :many
SELECT * FROM sub_projects WHERE user_id = ? and project_id = ? ORDER BY sub_project; 

-- name: SelectSubProjectIdBySubProjectName :one
SELECT id FROM sub_projects WHERE sub_project = ? AND user_id = ? AND project_id = ?;

-- name: InsertSubProject :one
INSERT INTO sub_projects (user_id, project_id, sub_project, description, status) VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateSubProjectsByUserAndProject :exec
UPDATE sub_projects SET sub_project = ?, description = ?, status = ? WHERE id = ? AND user_id = ? AND project_id = ?;

-- name: DeleteSubProjects :one
DELETE FROM sub_projects WHERE user_id = ? AND id = ? AND project_id = ? RETURNING *;
