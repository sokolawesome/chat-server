package handlers

import (
	"database/sql"
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
	UserRepository    repository.UserRepository
	JwtSecret         string
	JwtExpirationTime time.Duration
}

func NewAuthHandler(userRepository repository.UserRepository, jwtSecret string, jwtExpirationTime time.Duration) *AuthHandler {
	return &AuthHandler{
		UserRepository:    userRepository,
		JwtSecret:         jwtSecret,
		JwtExpirationTime: jwtExpirationTime,
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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body: " + err.Error()})
		return
	}

	existingUser, err := h.UserRepository.GetUserByUsername(ctx.Request.Context(), req.Username)
	if err != nil {
		log.Printf("error checking existing user during registration: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process registration"})
		return
	}
	if existingUser != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "User with username already exist"})
		return
	}

	newUser, err := h.UserRepository.CreateUser(ctx.Request.Context(), req.Username, req.Password)
	if err != nil {
		log.Printf("error creating user: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

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
	Token    string       `json:"token"`
	User     *models.User `json:"user"`
	Username string       `json:"username"`
}

func (h *AuthHandler) Login(ctx *gin.Context) {
	var req LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("login validation error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body: " + err.Error()})
		return
	}

	user, err := h.UserRepository.GetUserByUsername(ctx.Request.Context(), req.Username)
	if err != nil {
		if err == sql.ErrNoRows || user == nil {
			log.Printf("user with username '%s' not found", req.Username)
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		log.Printf("error fetching user during login: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed due to server error"})
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password)); err != nil {
		log.Printf("password for user '%s' is invalid", req.Username)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss": "chat-app",
		"sub": user.ID,
		"usr": user.Username,
		"iat": now.Unix(),
		"exp": now.Add(h.JwtExpirationTime).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenSigned, err := token.SignedString([]byte(h.JwtSecret))
	if err != nil {
		log.Printf("error signing jwt for user %s: %v", user.Username, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	log.Printf("User %d (%s) logged in successfully", user.ID, user.Username)

	respone := LoginResponse{
		Token:    tokenSigned,
		User:     user,
		Username: user.Username,
	}

	ctx.JSON(http.StatusOK, respone)
}
