package main

import (
	"Trames/cmd/config"
	logger "Trames/cmd/config"
	coredb "Trames/internal/core/db"
	"Trames/internal/core/tasks"
	"Trames/internal/core/users"
	"context"
	"database/sql"
	"os"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	appCtx := context.Background()
	setupCtx, cancel := context.WithTimeout(appCtx, 10*time.Second)
	defer cancel()

	logger := logger.New(os.Stdout, logger.LevelInfo)

	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Fatal("Failed to Load the configuration of the server: "+err.Error(), nil)
	}

	dbPool, err := sql.Open("sqlite", "./trames.db")
	if err != nil {
		logger.Fatal("Failed to open the db: "+err.Error(), nil)
	}
	defer dbPool.Close()
	if err := dbPool.PingContext(setupCtx); err != nil {
		logger.Fatal("Failed to ping the DB: "+err.Error(), nil)
	}

	userHandler := &users.UserHandle{
		DB:      dbPool,
		Queries: coredb.New(dbPool),
		Logger:  logger,
		SessionConfig: users.SessionConfig{
			SessionExpiration: cfg.Secret.SessionExpiration,
		},
	}
	handler := &tasks.Handler{
		DB:      dbPool,
		Queries: coredb.New(dbPool),
		Logger:  logger,
	}

	app := &config.App{
		UserHandler: userHandler,
		Handler:     handler,
		Wg:          &sync.WaitGroup{},
	}

	if err := app.Serve(logger, cfg); err != nil {
		logger.Fatal("server error: "+err.Error(), nil)
	}
	logger.Info("trames shutting down", nil)
}
