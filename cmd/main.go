package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"projeto_drm/poc/internal/auth"
	"projeto_drm/poc/internal/database"
	"projeto_drm/poc/internal/handlers"
	"time"
)

func main() {
	database.InitDatabase()

	r := gin.Default()
	r.POST("/auth/login", auth.LoginHandler)
	r.Use(auth.Middleware())

	handlers.RegisterRoutes(r)

	s := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    240 * time.Second,
		WriteTimeout:   120 * time.Second,
		IdleTimeout:    180 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
