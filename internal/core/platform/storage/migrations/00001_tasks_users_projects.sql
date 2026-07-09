-- +goose Up

CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_users_username_unique
ON users(username);


CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    project TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'TO BE STARTED'
        CHECK (status IN ('IDEA', 'TO BE STARTED', 'PLANNING', 'IN PROGRESS', 'DONE')),

    FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_projects_user_project_unique
ON projects(user_id, project);

CREATE UNIQUE INDEX idx_projects_user_id_id_unique
ON projects(user_id, id);


CREATE TABLE sub_projects (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    project_id INTEGER NOT NULL,
    sub_project TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'TO BE STARTED'
        CHECK (status IN ('IDEA', 'TO BE STARTED', 'PLANNING', 'IN PROGRESS', 'DONE')),

    FOREIGN KEY (user_id, project_id)
        REFERENCES projects(user_id, id)
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_sub_projects_user_project_sub_project_unique
ON sub_projects(user_id, project_id, sub_project);

CREATE UNIQUE INDEX idx_sub_projects_user_project_id_id_unique
ON sub_projects(user_id, project_id, id);


CREATE TABLE tasks (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    project_id INTEGER NOT NULL,
    sub_project_id INTEGER NULL,
    task TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'TO-DO'
        CHECK (status IN ('EXPERIMENTAL', 'TO-DO', 'IN PROGRESS', 'BLOCKED', 'TEST', 'DONE')),
    priority TEXT NOT NULL DEFAULT 'LOW'
        CHECK (priority IN ('IDEA', 'LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),

    FOREIGN KEY (user_id, project_id)
        REFERENCES projects(user_id, id)
        ON DELETE CASCADE,

    FOREIGN KEY (user_id, project_id, sub_project_id)
        REFERENCES sub_projects(user_id, project_id, id)
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_tasks_root_task_unique
ON tasks(user_id, project_id, task)
WHERE sub_project_id IS NULL;

CREATE UNIQUE INDEX idx_tasks_sub_project_task_unique
ON tasks(user_id, project_id, sub_project_id, task)
WHERE sub_project_id IS NOT NULL;

-- +goose StatementBegin 
CREATE TRIGGER users_updated_at 
AFTER UPDATE OF username, password_hash ON users 
FOR EACH ROW 
BEGIN 
    UPDATE users 
    SET updated_at = CURRENT_TIMESTAMP 
    WHERE id = OLD.id; 
END; 
-- +goose StatementEnd

-- +goose Down 
DROP TRIGGER IF EXISTS users_updated_at;
DROP TABLE IF EXISTS tasks; 
DROP TABLE IF EXISTS sub_projects; 
DROP TABLE IF EXISTS projects; 
DROP TABLE IF EXISTS users;
