package config

import (
	"log"
	"os"
)

type Config struct {
	ServerPort  string
	DatabaseUrl string
}

func Load() (*Config, error) {
	cfg := Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		DatabaseUrl: getEnv("DATABASE_URL", ""),
	}

	if cfg.DatabaseUrl == "" {
		log.Fatal("error: DATABASE_URL environment variable is required")
	}

	return &cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
