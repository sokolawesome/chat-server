package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sokolawesome/chat-server/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUsernameTaken   = errors.New("username is already taken")
	ErrHashingPassword = errors.New("failed to hash password")
	ErrCreatingUser    = errors.New("failed to create user in database")
	ErrRetrievingUser  = errors.New("failed to retrieve user from database")
)

type UserRepository interface {
	CreateUser(ctx context.Context, username string, password string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}

type postgresUserRepository struct {
	db         *sql.DB
	bcryptCost int
}

func NewUserRepository(db *sql.DB, bcryptCost int) UserRepository {
	return &postgresUserRepository{db: db, bcryptCost: bcryptCost}
}

func (r *postgresUserRepository) CreateUser(ctx context.Context, username string, password string) (*models.User, error) {
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), r.bcryptCost)
	if err != nil {
		log.Printf("error hashing password for user %s: %v", username, err)
		return nil, fmt.Errorf("%w: %v", ErrHashingPassword, err)
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			log.Printf("attempt to create user with existing username '%s'", username)
			return nil, ErrUsernameTaken
		}
		log.Printf("error inserting user '%s' into database: %v", username, err)
		return nil, fmt.Errorf("%w: %v", ErrCreatingUser, err)
	}

	log.Printf("user created successfully with id: %d, username: %s", user.ID, user.Username)
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
			log.Printf("user not found by username '%s'", username)
			return nil, ErrUserNotFound
		}
		log.Printf("error retrieving user by username '%s' from database: %v", username, err)
		return nil, fmt.Errorf("%w: %v", ErrRetrievingUser, err)
	}

	return user, nil
}
