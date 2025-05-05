package config

import (
	"log"
	"os"
)

type Config struct {
	ServerPort  string
	DatabaseUrl string
	JwtSecret   string
}

func Load() (*Config, error) {
	cfg := Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		DatabaseUrl: getEnv("DATABASE_URL", ""),
		JwtSecret:   getEnv("JWT_SECRET", ""),
	}

	if cfg.DatabaseUrl == "" {
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
