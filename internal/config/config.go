package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// Application
	Environment string
	Port        string
	ClientURL   string

	// Database
	DatabaseURL     string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifeTime time.Duration
	ConnMaxIdleTime time.Duration

	// Authentication
	JWTSecret    string
	JWTExpiresIn time.Duration

	// Storage
	StorageType   string // "local" or "s3"
	LocalBasePath string
	S3Region      string
	S3Bucket      string
	S3AccessKey   string
	S3SecretKey   string

	// Giphy
	GiphyApiKey string

	// Unsplash
	UnsplashAccessKey string
}

// Load returns application configuration from environment variables
func Load() *Config {
	// JWT expiration in hours, default to 24 if not specified
	jwtExpiration, _ := strconv.Atoi(getEnv("JWT_EXPIRES_IN", "24"))

	maxIdleConn, _ := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "10"))
	maxOpenConn, _ := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "100"))
	connMaxLifetime, _ := strconv.Atoi(getEnv("DB_CONN_MAX_LIFETIME", "60"))
	connMaxIdleTime, _ := strconv.Atoi(getEnv("DB_CONN_MAX_IDLE_TIME", "30"))

	return &Config{
		// Application config
		Environment: getEnv("APP_ENV", "development"),
		Port:        getEnv("PORT", "8080"),
		ClientURL:   getEnv("CLIENT_URL", "http://localhost:3000"),

		// Database
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/kudoboard?sslmode=disable"),
		MaxIdleConns:    maxIdleConn,
		MaxOpenConns:    maxOpenConn,
		ConnMaxLifeTime: time.Duration(connMaxLifetime) * time.Minute,
		ConnMaxIdleTime: time.Duration(connMaxIdleTime) * time.Minute,

		// Authentication
		JWTSecret:    getEnv("JWT_SECRET", "your-super-secret-key-change-this-in-production"),
		JWTExpiresIn: time.Duration(jwtExpiration) * time.Hour,

		// Storage
		StorageType:   getEnv("STORAGE_TYPE", "local"),
		LocalBasePath: getEnv("LOCAL_STORAGE_PATH", "./uploads"),
		S3Region:      getEnv("S3_REGION", ""),
		S3Bucket:      getEnv("S3_BUCKET", ""),
		S3AccessKey:   getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:   getEnv("S3_SECRET_KEY", ""),

		// Giphy
		GiphyApiKey: getEnv("GIPHY_API_KEY", ""),

		// Unsplash
		UnsplashAccessKey: getEnv("UNSPLASH_ACCESS_KEY", ""),
	}
}

// Helper function to get environment variables with a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
