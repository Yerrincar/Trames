package main

import (
	"Trames/cmd/config"
	logger "Trames/cmd/config"
	"database/sql"
	"os"
	"sync"
)

func main() {
	logger := logger.New(os.Stdout, logger.LevelInfo)

	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Fatal("Failed to Load the configuration of the server: "+err.Error(), nil)
	}

	db, err := sql.Open("sqlite", "./trames.db")
	if err != nil {
		logger.Fatal("Failed to open the db: "+err.Error(), nil)
	}

	app := &config.App{
		Wg: &sync.WaitGroup{},
	}

	app.Serve(logger, cfg)
	logger.Info("librorum shutting down", nil)
}
