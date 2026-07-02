-- +goose Up 
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    task TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'TO-DO' CHECK (status IN ('EXPERIMENTAL', 'TO-DO', 'IN PROGRESS', 'BLOCKED', 'TEST', 'DONE')),
    priority TEXT NOT NULL DEFAULT 'LOW' CHECK (priority IN ('IDEA','LOW', 'MEDIUM', 'HIGH', 'CRITICAL'))
);

CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'TO BE STARTED' CHECK (status IN('IDEA', 'TO BE STARTED', 'PLANNING', 'IN PROGRESS', 'DONE'))
);

CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sessions (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TEXT NOT NULL
);


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
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS sessions;
DROP TRIGGER IF EXISTS users_updated_at;
