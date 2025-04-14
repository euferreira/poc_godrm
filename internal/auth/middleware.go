package auth

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"net/http"
	"os"
	"strings"
)

type UserInfo struct {
	ID    string `json:"userID"`
	Email string `json:"email"`
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil {
			var ve *jwt.ValidationError
			if errors.As(err, &ve) && ve.Errors&jwt.ValidationErrorExpired != 0 {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expirado"})
				c.Abort()
				return
			}

			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			user := UserInfo{
				ID:    fmt.Sprintf("%v", claims["userID"]),
				Email: fmt.Sprintf("%v", claims["email"]),
			}
			c.Set("user", user)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			c.Abort()
			return
		}

		c.Next()
	}
}
