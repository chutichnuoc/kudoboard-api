package db

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/models"
	"log"
	"time"
)

// Connect establishes connection to the database
func Connect(cfg *config.Config) (*gorm.DB, error) {
	// Configure GORM logger
	gormLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  getLogLevel(cfg.Environment),
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool parameters
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Connected to database")
	return db, nil
}

// MigrateSchema performs database migrations
func MigrateSchema(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Auto-migrate all models
	err := db.AutoMigrate(
		&models.User{},
		&models.Theme{},
		&models.Board{},
		&models.BoardContributor{},
		&models.Post{},
		&models.Media{},
		&models.PostLike{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Create default themes if they don't exist
	createDefaultThemes(db)

	log.Println("Database migrations completed")
	return nil
}

// Create default themes
func createDefaultThemes(db *gorm.DB) {
	var count int64
	db.Model(&models.Theme{}).Count(&count)

	// Only create default themes if none exist
	if count == 0 {
		themes := []models.Theme{
			{
				Name:               "Default Light",
				Description:        "Clean light theme with subtle background",
				BackgroundColor:    "#ffffff",
				BackgroundImageURL: "",
				AdditionalStyles:   `{"cardBgColor": "#f8f9fa", "textColor": "#212529"}`,
				IsDefault:          true,
			},
			{
				Name:               "Default Dark",
				Description:        "Modern dark theme",
				BackgroundColor:    "#212529",
				BackgroundImageURL: "",
				AdditionalStyles:   `{"cardBgColor": "#343a40", "textColor": "#f8f9fa"}`,
				IsDefault:          false,
			},
			{
				Name:               "Celebration",
				Description:        "Festive theme with confetti background",
				BackgroundColor:    "#f8f9fa",
				BackgroundImageURL: "/themes/celebration-bg.jpg",
				AdditionalStyles:   `{"cardBgColor": "#ffffff", "textColor": "#212529"}`,
				IsDefault:          false,
			},
			{
				Name:               "Gratitude",
				Description:        "Warm theme perfect for thank you messages",
				BackgroundColor:    "#fff8e1",
				BackgroundImageURL: "",
				AdditionalStyles:   `{"cardBgColor": "#fffde7", "textColor": "#5d4037"}`,
				IsDefault:          false,
			},
			{
				Name:               "Professional",
				Description:        "Clean and formal theme for workplace kudos",
				BackgroundColor:    "#eceff1",
				BackgroundImageURL: "",
				AdditionalStyles:   `{"cardBgColor": "#ffffff", "textColor": "#263238"}`,
				IsDefault:          false,
			},
		}

		for _, theme := range themes {
			db.Create(&theme)
		}
		log.Println("Created default themes")
	}
}

// getLogLevel returns the appropriate GORM log level based on environment
func getLogLevel(env string) logger.LogLevel {
	if env == "production" {
		return logger.Error
	}
	return logger.Info
}
