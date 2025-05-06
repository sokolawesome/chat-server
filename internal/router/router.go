package router

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sokolawesome/chat-server/config"
	"github.com/sokolawesome/chat-server/internal/handlers"
	"github.com/sokolawesome/chat-server/internal/middleware"
)

func SetupRouter(cfg *config.Config, AuthHandler *handlers.AuthHandler) *gin.Engine {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Chat server is running!",
		})
	})

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", AuthHandler.Register)
			auth.POST("/login", AuthHandler.Login)
		}

		authorized := api.Group("/")
		authorized.Use(middleware.AuthMiddleware(cfg.JwtSecret))
		{
			authorized.GET("/me", func(ctx *gin.Context) {
				userIDAny, exist := ctx.Get(middleware.AuthorizationPayloadKey)
				if !exist {
					log.Println("userid not found in context")
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not identify user"})
					return
				}

				userID, ok := userIDAny.(int64)
				if !ok {
					log.Printf("userid in context is not int64 (%T)", userIDAny)
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not identify user"})
					return
				}

				log.Printf("/me endpoint accessed by user %d", userID)
				ctx.JSON(http.StatusOK, gin.H{
					"message": "Authentication successful",
					"user_id": userID,
				})
			})
		}
	}

	return router
}
