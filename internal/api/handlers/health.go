package handlers

import (
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/dto/responses"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Version will be set during build
var Version = "dev"

// HealthHandler handles health check requests
type HealthHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(db *gorm.DB, cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		db:  db,
		cfg: cfg,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      string            `json:"status"`
	Version     string            `json:"version"`
	Environment string            `json:"environment"`
	Timestamp   time.Time         `json:"timestamp"`
	Components  map[string]string `json:"components,omitempty"`
	Uptime      string            `json:"uptime,omitempty"`
}

var startTime = time.Now()

// LivenessCheck handles liveness probe requests
// This is a simple check to verify the application is running
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, responses.SuccessResponse(HealthResponse{
		Status:      "UP",
		Version:     Version,
		Environment: h.cfg.Environment,
		Timestamp:   time.Now(),
		Uptime:      time.Since(startTime).String(),
	}))
}

// ReadinessCheck handles readiness probe requests
// This checks if the application is ready to handle requests
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	components := make(map[string]string)
	status := "UP"

	// Check database connection
	sqlDB, err := h.db.DB()
	if err != nil {
		components["database"] = "DOWN: " + err.Error()
		status = "DOWN"
	} else if err = sqlDB.Ping(); err != nil {
		components["database"] = "DOWN: " + err.Error()
		status = "DOWN"
	} else {
		components["database"] = "UP"
	}

	// Check S3 connection if configured
	if h.cfg.StorageType == "s3" {
		// We could add code to check S3 connectivity here
		components["storage"] = "UP" // Simplified for now
	}

	// You could add checks for other dependencies like Redis, external APIs, etc.
	if h.cfg.GiphyApiKey != "" {
		components["giphy"] = "CONFIGURED"
	}

	if h.cfg.UnsplashAccessKey != "" {
		components["unsplash"] = "CONFIGURED"
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(HealthResponse{
		Status:      status,
		Version:     Version,
		Environment: h.cfg.Environment,
		Timestamp:   time.Now(),
		Components:  components,
		Uptime:      time.Since(startTime).String(),
	}))
}

// DetailedHealthCheck provides a comprehensive health check
func (h *HealthHandler) DetailedHealthCheck(c *gin.Context) {
	components := make(map[string]string)
	status := "UP"

	// Check database connection and get stats
	sqlDB, err := h.db.DB()
	if err != nil {
		components["database"] = "DOWN: " + err.Error()
		status = "DOWN"
	} else if err = sqlDB.Ping(); err != nil {
		components["database"] = "DOWN: " + err.Error()
		status = "DOWN"
	} else {
		stats := sqlDB.Stats()
		components["database"] = "UP"
		components["database_open_connections"] = strconv.Itoa(stats.OpenConnections)
		components["database_in_use"] = strconv.Itoa(stats.InUse)
		components["database_idle"] = strconv.Itoa(stats.Idle)
	}

	// Add more detailed checks for other components
	// This could include checking external service response times, etc.

	response := HealthResponse{
		Status:      status,
		Version:     Version,
		Environment: h.cfg.Environment,
		Timestamp:   time.Now(),
		Components:  components,
		Uptime:      time.Since(startTime).String(),
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(response))
}
