package config

import (
	"fmt"
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
	JwtIssuer             string
	DbMaxOpenConns        int
	DbMaxIdleConns        int
	DbConnMaxLifetime     time.Duration
	BcryptCost            int
	WsReadBufferSize      int
	WsWriteBufferSize     int
}

func Load() (*Config, error) {
	serverPort := getEnv("SERVER_PORT", "8080")
	databaseURL := getEnv("DATABASE_URL", "")
	jwtSecret := getEnv("JWT_SECRET", "")
	jwtIssuer := getEnv("JWT_ISSUER", "chat-app")
	jwtExpirationDuration, err := time.ParseDuration(getEnv("JWT_EXPIRATION_DURATION", "1h"))
	if err != nil {
		log.Printf("warning: could not parse JWT_EXPIRATION_DURATION '%s', using default 1h: %v", jwtExpirationDuration, err)
		jwtExpirationDuration = time.Hour
	}
	dbMaxOpenConns := getEnvAsInt("DB_MAX_OPEN_CONNS", 10)
	dbMaxIdleConns := getEnvAsInt("DB_MAX_IDLE_CONNS", 10)
	dbConnMaxLifetime, err := time.ParseDuration(getEnv("DB_CONN_MAX_LIFETIME", "5m"))
	if err != nil {
		log.Printf("warning: could not parse DB_CONN_MAX_LIFETIME '%s', using default 5m: %v", dbConnMaxLifetime, err)
		dbConnMaxLifetime = 5 * time.Minute
	}
	bcryptCost := getEnvAsInt("BCRYPT_COST", 12)
	wsReadBufferSize := getEnvAsInt("WS_READ_BUFFER_SIZE", 1024)
	wsWriteBufferSize := getEnvAsInt("WS_WRITE_BUFFER_SIZE", 1024)

	cfg := &Config{
		ServerPort:            serverPort,
		DatabaseURL:           databaseURL,
		JwtSecret:             jwtSecret,
		JwtExpirationDuration: jwtExpirationDuration,
		JwtIssuer:             jwtIssuer,
		DbMaxOpenConns:        dbMaxOpenConns,
		DbMaxIdleConns:        dbMaxIdleConns,
		DbConnMaxLifetime:     dbConnMaxLifetime,
		BcryptCost:            bcryptCost,
		WsReadBufferSize:      wsReadBufferSize,
		WsWriteBufferSize:     wsWriteBufferSize,
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("config error: DATABASE_URL environment variable is required")
	}
	if cfg.JwtSecret == "" {
		return nil, fmt.Errorf("config error: JWT_SECRET environment variable is required")
	}
	if cfg.BcryptCost < 4 || cfg.BcryptCost > 31 {
		return nil, fmt.Errorf("config error: BCRYPT_COST must be between 4 and 31, got %d", cfg.BcryptCost)
	}

	log.Println("[Config] Loaded successfully.")

	return cfg, nil
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
