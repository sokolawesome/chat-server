package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sokolawesome/chat-server/config"
)

func Connect(cfg *config.Config) (*sql.DB, error) {
	log.Println("connecting to database...")

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("database.Connect: failed to open connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.DbMaxOpenConns)
	db.SetMaxIdleConns(cfg.DbMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DbConnMaxLifetime)

	if err = db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("database.Connect: error closing connection after ping failed: %v", closeErr)
		}
		return nil, fmt.Errorf("database.Connect: failed to ping database: %w", err)
	}

	log.Println("database connection established successfully")
	return db, nil
}
