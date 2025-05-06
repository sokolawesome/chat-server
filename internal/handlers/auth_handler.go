package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sokolawesome/chat-server/internal/models"
	"github.com/sokolawesome/chat-server/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	UserRepository        repository.UserRepository
	JwtSecret             string
	JwtExpirationDuration time.Duration
	JwtIssuer             string
}

func NewAuthHandler(userRepository repository.UserRepository, jwtSecret string, jwtExpirationDuration time.Duration, jwtIssuer string) *AuthHandler {
	return &AuthHandler{
		UserRepository:        userRepository,
		JwtSecret:             jwtSecret,
		JwtExpirationDuration: jwtExpirationDuration,
		JwtIssuer:             jwtIssuer,
	}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *AuthHandler) Register(ctx *gin.Context) {
	var req RegisterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("register validation error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	_, err := h.UserRepository.GetUserByUsername(ctx.Request.Context(), req.Username)
	if err == nil {
		log.Printf("registration attempt with existing username: %s", req.Username)
		ctx.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
		return
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		log.Printf("error checking existing user '%s' during registration: %v", req.Username, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process registration"})
		return
	}

	newUser, err := h.UserRepository.CreateUser(ctx.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrUsernameTaken) {
			log.Printf("failed to create user, username '%s' already taken: %v", req.Username, err)
			ctx.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
			return
		}
		log.Printf("error creating user '%s': %v", req.Username, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	log.Printf("user registered successfully: id=%d, username=%s", newUser.ID, newUser.Username)
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user_id": newUser.ID,
	})
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (h *AuthHandler) Login(ctx *gin.Context) {
	var req LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("login validation error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	user, err := h.UserRepository.GetUserByUsername(ctx.Request.Context(), req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Printf("login attempt for non-existent username: %s", req.Username)
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		log.Printf("error fetching user '%s' during login: %v", req.Username, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed due to server error"})
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password)); err != nil {
		log.Printf("invalid password attempt for username: %s", req.Username)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss": h.JwtIssuer,
		"sub": user.ID,
		"usr": user.Username,
		"iat": now.Unix(),
		"exp": now.Add(h.JwtExpirationDuration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenSigned, err := token.SignedString([]byte(h.JwtSecret))
	if err != nil {
		log.Printf("error signing jwt for user %s: %v", user.Username, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	log.Printf("User %d (%s) logged in successfully", user.ID, user.Username)

	response := LoginResponse{
		Token: tokenSigned,
		User:  user,
	}

	ctx.JSON(http.StatusOK, response)
}
