package main

import (
	"context"
	"errors"
	"kudoboard-api/internal/services"
	"kudoboard-api/internal/services/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"kudoboard-api/internal/api/routes"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/db"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
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
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate database schema
	if err := db.MigrateSchema(database); err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
	}

	// Initialize storage service
	storageService, err := storage.NewStorageService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage service: %v", err)
	}

	// Initialize services
	authService := services.NewAuthService(database, cfg)
	boardService := services.NewBoardService(database, storageService, cfg)
	postService := services.NewPostService(database, storageService, cfg, boardService)
	mediaService := services.NewMediaService(database, storageService, cfg, boardService)

	// Create Gin router
	router := gin.Default()

	// Setup routes with all services
	routes.Setup(
		router,
		database,
		cfg,
		authService,
		boardService,
		postService,
		mediaService,
	)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server running on port %s\n", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
