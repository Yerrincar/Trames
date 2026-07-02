package tasks

import (
	"Trames/internal/core/db"
	"database/sql"
)

type Logger interface {
	Info(message string, properties map[string]string)
	Error(message string, properties map[string]string)
	Debug(message string, properties map[string]string)
}

type Handler struct {
	Queries *db.Queries
	DB      *sql.DB
	Logger  Logger
}

type TaskForm struct {
	Task        string
	Description sql.NullString
	Status      string
	Priority    string
}
