package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	AuthorizationHeaderKey  = "Authorization"
	AuthorizationTypeBearer = "bearer"
	AuthorizationPayloadKey = "authorization_payload"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(AuthorizationHeaderKey)

		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			log.Println("auth error: ", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) != 2 {
			err := errors.New("invalid authorization header format")
			log.Println("auth error: ", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != AuthorizationTypeBearer {
			err := fmt.Errorf("unsupported auth type %s", authorizationType)
			log.Println("auth error: ", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		accessToken := fields[1]
		token, err := jwt.Parse(accessToken, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			log.Printf("jwt parsing/validation error: %v", err)
			errMsg := "invalid token"
			if errors.Is(err, jwt.ErrTokenExpired) {
				errMsg = "token has expired"
			}
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errMsg})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userIdF64, okSub := claims["sub"].(float64)
			username, okUsr := claims["usr"].(string)
			if !okSub {
				err := errors.New("invalid token: missing or invalid userid (sub) claim")
				log.Println("auth error:", err)
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token payload"})
				return
			}
			if !okUsr {
				err := errors.New("invalid token: missing or invalid username (usr) claim")
				log.Println("auth error:", err)
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token payload"})
				return
			}

			userId := int64(userIdF64)

			ctx.Set(AuthorizationPayloadKey, userId)

			log.Printf("auth success: user %d (%s) authorized", userId, username)

			ctx.Next()
		} else {
			log.Println("auth error: invalid token (claims invalid or token marked invalid)")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		}
	}
}
