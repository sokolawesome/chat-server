package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/sokolawesome/chat-server/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	CreateUser(ctx context.Context, username string, password string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}

type postgresUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) CreateUser(ctx context.Context, username string, password string) (*models.User, error) {
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("error hashing password: %v", err)
		return nil, err
	}

	hashedPassword := string(hashedPasswordBytes)

	user := &models.User{
		Username:       username,
		HashedPassword: hashedPassword,
	}

	query := `INSERT INTO users (username, hashed_password)
    VALUES ($1, $2)
    RETURNING id, created_at`

	if err = r.db.QueryRowContext(ctx, query, username, hashedPassword).Scan(&user.ID, &user.CreatedAt); err != nil {
		// check db errors later
		log.Printf("error inserting user: %v", err)
		return nil, err
	}

	log.Printf("user created successfully with id: %d", user.ID)
	return user, nil
}

func (r *postgresUserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, hashed_password, created_at
    FROM users
    WHERE username = $1`

	if err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.HashedPassword,
		&user.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		log.Printf("error retrieving user by username '%s': %v", username, err)
		return nil, err
	}

	return user, nil
}
