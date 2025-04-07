package main

import (
	"context"
	"errors"
	"kudoboard-api/internal/api/routes"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/container"
	"kudoboard-api/internal/db"
	"kudoboard-api/internal/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Warn("Warning: .env file not found")
	}

	// Initialize configuration
	cfg := config.Load()

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Migrate database schema
	if err := db.MigrateSchema(database); err != nil {
		log.Fatal("Failed to migrate database schema", zap.Error(err))
	}

	// Create service container
	serviceContainer, err := container.NewContainer(cfg, database)
	if err != nil {
		log.Fatal("Failed to initialize service container", zap.Error(err))
	}

	// Create Gin router
	router := gin.New()

	// Setup routes with the container
	routes.Setup(router, cfg, serviceContainer)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Info("Server running", zap.String("port", cfg.Port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	// Flush any buffered log entries
	log.Shutdown()

	log.Info("Server exited properly")
}
