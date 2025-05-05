package router

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sokolawesome/chat-server/internal/handlers"
)

func SetupRouter(AuthHandler *handlers.AuthHandler) *gin.Engine {
	router := gin.Default()

	cfg := cors.DefaultConfig()
	cfg.AllowOrigins = []string{"*"}
	cfg.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	cfg.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	router.Use(cors.New(cfg))

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
	}

	return router
}
