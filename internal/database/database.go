package database

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sokolawesome/chat-server/config"
)

func Connect(cfg *config.Config) (*sql.DB, error) {
	log.Println("connecting to database...")

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.DbMaxOpenConns)
	db.SetMaxIdleConns(cfg.DbMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DbConnMaxLifetime)

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	log.Println("database connection established successfully")
	return db, nil
}
