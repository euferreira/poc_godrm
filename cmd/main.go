package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"projeto_drm/poc/internal/auth"
	"projeto_drm/poc/internal/cleanup"
	"projeto_drm/poc/internal/database"
	"projeto_drm/poc/internal/handlers"
	"projeto_drm/poc/internal/queue"
	"projeto_drm/poc/internal/worker"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	database.InitDatabase()

	r := gin.Default()
	r.POST("/auth/login", auth.LoginHandler)
	r.Use(auth.Middleware())

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis:6379"
	}

	handlers.RegisterRoutes(r)
	queue.InitRedisClient()
	handlers.InitializeQueue(redisURL)
	redisQueue := queue.NewRedisQueue(redisURL)

	// Inicializar worker pools
	workerPool := worker.NewWorkerPool(3, redisQueue) // 3 workers
	workerPool.Start()

	// Inicializar file copy worker pool
	fileCopyWorkerPool := worker.NewFileCopyWorkerPool(2) // 2 workers
	fileCopyWorkerPool.Start()

	// Inicializar cleanup automático (limpa arquivos com mais de 24 horas)
	cleanup.StartCacheCleanup(time.Hour, 24*time.Hour)

	s := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    240 * time.Second,
		WriteTimeout:   120 * time.Second,
		IdleTimeout:    180 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Shutting down...")
		workerPool.Stop()
		fileCopyWorkerPool.Stop()
		os.Exit(0)
	}()

	log.Fatal(s.ListenAndServe())
}
