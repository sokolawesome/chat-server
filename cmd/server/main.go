package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/sokolawesome/chat-server/config"
	"github.com/sokolawesome/chat-server/internal/database"
	"github.com/sokolawesome/chat-server/internal/handlers"
	"github.com/sokolawesome/chat-server/internal/repository"
	"github.com/sokolawesome/chat-server/internal/router"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// proper origin check later
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleWebSocket(ctx *gin.Context, cfg *config.Config) {
	tokenString := ctx.Query("token")
	if tokenString == "" {
		log.Println("missing token in query parameters")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authentication token"})
		ctx.Abort()
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JwtSecret), nil
	})

	if err != nil {
		log.Printf("jwt parsing/validation error: %v", err)
		errMsg := "invalid or expired token"
		if errors.Is(err, jwt.ErrTokenExpired) {
			errMsg = "token has expired"
		}
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": errMsg})
		ctx.Abort()
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDF64, okSub := claims["sub"].(float64)
		if !okSub {
			log.Println("invalid token payload (missing/invalid sub claim)")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token payload"})
			ctx.Abort()
			return
		}
		userID := int64(userIDF64)
		log.Printf("user %d authorized for websocket connection", userID)
	} else {
		log.Println("invalid token (claims invalid or token marked invalid)")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		ctx.Abort()
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("failed to upgrade connection from %s: %v", ctx.Request.RemoteAddr, err)
		return
	}
	defer func() {
		log.Println("closing connection for client:", conn.RemoteAddr())
		if err := conn.Close(); err != nil {
			log.Printf("error closing websocket connection for %s: %v", conn.RemoteAddr(), err)
		} else {
			log.Println("websocket connection closed successfully for:", conn.RemoteAddr())
		}
	}()

	log.Println("websocket client connected:", conn.RemoteAddr())

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket error for client %s: %v", conn.RemoteAddr(), err)
			} else {
				log.Println("websocket client disconnected:", conn.RemoteAddr())
			}
			break
		}

		log.Printf("received message (type %d) from %s: %s", messageType, conn.RemoteAddr(), string(p))

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Printf("failed to write message to client %s: %v", conn.RemoteAddr(), err)
			break
		}
	}

	log.Printf("handler finished for client: %s", conn.RemoteAddr())
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	defer func() {
		log.Println("closing database connection pool...")
		if err := db.Close(); err != nil {
			log.Printf("error closing database connection: %v", err)
		}
	}()

	userRepository := repository.NewUserRepository(db)
	authHandler := handlers.NewAuthHandler(userRepository, cfg.JwtSecret, cfg.JwtExpirationDuration)
	ginRouter := router.SetupRouter(cfg, authHandler)

	ginRouter.GET("/ws", func(ctx *gin.Context) {
		handleWebSocket(ctx, cfg)
	})

	log.Printf("server listening on http://localhost:%s", cfg.ServerPort)
	if err := ginRouter.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
