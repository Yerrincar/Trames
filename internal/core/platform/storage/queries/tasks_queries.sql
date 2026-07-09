-- name: SelectTasksByUserAndProject :many
SELECT
    tasks.id,
    tasks.task,
    tasks.description,
    tasks.status,
    tasks.priority,
    tasks.project_id,
    tasks.sub_project_id,
    sub_projects.sub_project
FROM tasks
LEFT JOIN sub_projects
    ON sub_projects.id = tasks.sub_project_id
   AND sub_projects.user_id = tasks.user_id
   AND sub_projects.project_id = tasks.project_id
WHERE tasks.user_id = ?
  AND tasks.project_id = ?
ORDER BY tasks.id DESC;

-- name: SelectTasksByUserAndProjectAndSubProject :many
SELECT
    tasks.id,
    tasks.task,
    tasks.description,
    tasks.status,
    tasks.priority,
    tasks.project_id,
    tasks.sub_project_id,
    sub_projects.sub_project
FROM tasks
LEFT JOIN sub_projects
    ON sub_projects.id = tasks.sub_project_id
   AND sub_projects.user_id = tasks.user_id
   AND sub_projects.project_id = tasks.project_id
WHERE tasks.user_id = ?
  AND tasks.project_id = ?
  AND tasks.sub_project_id = ?
ORDER BY tasks.id DESC;

-- name: InsertTasksByUserProjectAndSubProject :one
INSERT INTO tasks (user_id, sub_project_id, project_id, task, description, status, priority) VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: InsertTasksByUserAndProject :one
INSERT INTO tasks (user_id, sub_project_id, project_id, task, description, status, priority) VALUES (?, NULL, ?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateTasksByUserAndProject :exec
UPDATE tasks SET task = ?, description = ?, status = ?, priority = ? WHERE id = ? AND user_id = ?;

-- name: DeleteTask :many
DELETE FROM tasks WHERE id IN (sqlc.slice('ids')) RETURNING *;
