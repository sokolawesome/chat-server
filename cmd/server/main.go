package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// proper origin check later
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleWebSocket(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("failded to upgrade connection from %s: %v", ctx.Request.RemoteAddr, err)
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
	if err := godotenv.Load(); err != nil {
		log.Printf(".env file not loaded: %v", err)
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("starting server on port", port)

	router := gin.Default()

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Chat server is running!",
		})
	})

	router.GET("/ws", handleWebSocket)

	log.Printf("server listening on http://localhost:%s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
