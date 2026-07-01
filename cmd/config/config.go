package config

import (
	"flag"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr            string
	TokenExpiration struct {
		durationString string
		duration       time.Duration
	}
	Secret struct {
		SessionExpiration time.Duration
	}
}

func LoadConfig(l *Logger) (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		l.Fatal("Error trying to Load .env: "+err.Error(), nil)
	}
	if err != nil {
		l.Fatal("Error trying to Read DB_MAX_IDLE_CONNS from .env %v"+err.Error(), nil)
	}
	var cfg Config

	flag.StringVar(&cfg.Addr, "addr", os.Getenv("TRAMES_ADDR"), "Address")
	sessionDuration, err := time.ParseDuration(os.Getenv("SESSION_EXPIRATION"))
	if err != nil {
		return nil, err
	}
	cfg.Secret.SessionExpiration = sessionDuration

	tokexpirationStr := os.Getenv("TOKEN_EXPIRATION")
	duration, err := time.ParseDuration(tokexpirationStr)
	if err != nil {
		return nil, err
	}
	cfg.TokenExpiration.durationString = tokexpirationStr
	cfg.TokenExpiration.duration = duration
	flag.Parse()

	return &cfg, nil
}
