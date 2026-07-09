package main

import (
	"Trames/cmd/config"
	logger "Trames/cmd/config"
	coredb "Trames/internal/core/db"
	"Trames/internal/core/operations"
	"Trames/internal/core/users"
	"context"
	"database/sql"
	"fmt"
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

	dbPool, err := sql.Open("sqlite", "file:./trames.db?_pragma=foreign_keys(1)")
	if err != nil {
		logger.Fatal("Failed to open the db: "+err.Error(), nil)
	}

	if err := dbPool.Ping(); err != nil {
		logger.Fatal("Failed to ping the db: "+err.Error(), nil)
	}

	if err := verifyForeignKeys(dbPool); err != nil {
		logger.Fatal("Failed to verify FKs: "+err.Error(), nil)
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
	handler := &operations.Handler{
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

func verifyForeignKeys(db *sql.DB) error {
	var enabled int

	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&enabled); err != nil {
		return err
	}

	if enabled != 1 {
		return fmt.Errorf("sqlite foreign keys are not enabled")
	}

	return nil
}
