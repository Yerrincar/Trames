package users

import (
	"Trames/internal/core/db"
	"database/sql"
	"time"
)

type Logger interface {
	Info(message string, properties map[string]string)
	Error(message string, properties map[string]string)
	Debug(message string, properties map[string]string)
}

type SessionConfig struct {
	SessionExpiration time.Duration
}

type UserProfile struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type password struct {
	Plaintext *string
	Hash      string
}

type UserHandle struct {
	DB            *sql.DB
	Queries       *db.Queries
	Logger        Logger
	SessionConfig SessionConfig
}
