package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort            string
	DatabaseURL           string
	JwtSecret             string
	JwtExpirationDuration time.Duration
}

func Load() (*Config, error) {
	cfg := Config{
		ServerPort:            getEnv("SERVER_PORT", "8080"),
		DatabaseURL:           getEnv("DATABASE_URL", ""),
		JwtSecret:             getEnv("JWT_SECRET", ""),
		JwtExpirationDuration: time.Duration(getEnvAsInt("JWT_EXPIRATION_MINUTES", 60)) * time.Minute,
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("fatal: DATABASE_URL environment variable is required")
	}
	if cfg.JwtSecret == "" {
		log.Fatal("fatal: JWT_SECRET environment variable is required")
	}

	return &cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if value, err := strconv.Atoi(value); err == nil {
			return value
		}
	}
	return fallback
}
