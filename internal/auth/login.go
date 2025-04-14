package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"net/http"
	"time"
)

func LoginHandler(c *gin.Context) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": "12345",
		"email":  "user@example.com",
		"exp":    time.Now().Add(time.Hour * 24).Unix(), // Expira em 24 horas
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Não foi possível gerar o token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
