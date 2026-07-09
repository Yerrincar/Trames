package operations

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

type OperationForm struct {
	ID          int64
	Operation   string
	SubProject  string
	Description sql.NullString
	Status      string
	Priority    string
}
