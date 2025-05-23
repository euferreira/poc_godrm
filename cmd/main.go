package main

import (
	"log"
	"net/http"
	"projeto_drm/poc/internal/auth"
	"projeto_drm/poc/internal/database"
	"projeto_drm/poc/internal/handlers"
	"projeto_drm/poc/internal/queue"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	database.InitDatabase()

	r := gin.Default()
	r.POST("/auth/login", auth.LoginHandler)
	r.Use(auth.Middleware())

	handlers.RegisterRoutes(r)
	queue.InitRedisClient()
	queue.StartWorker()

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
