package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sokolawesome/chat-server/internal/repository"
)

type AuthHandler struct {
	UserRepository repository.UserRepository
}

func NewAuthHandler(userRepository repository.UserRepository) *AuthHandler {
	return &AuthHandler{UserRepository: userRepository}
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

func (h *AuthHandler) Login(ctx *gin.Context) {
	var req LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("login validation error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body: " + err.Error()})
		return
	}

	user, err := h.UserRepository.GetUserByUsername(ctx.Request.Context(), req.Username)
	if err != nil {
		log.Printf("error fetching user during login: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		return
	}

	if user == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Login checked successfully",
		"user_id": user.ID,
	})
}
