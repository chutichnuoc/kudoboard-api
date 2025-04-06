package db

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/log"
	"kudoboard-api/internal/models"
	"time"

	"go.uber.org/zap"
)

// ZapGormLogger implements gorm's logger interface using Zap logger
type ZapGormLogger struct {
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	LogLevel                  gormlogger.LogLevel
}

// NewZapGormLogger creates a new GORM logger that uses Zap
func NewZapGormLogger(slowThreshold time.Duration, ignoreRecordNotFound bool) *ZapGormLogger {
	return &ZapGormLogger{
		SlowThreshold:             slowThreshold,
		IgnoreRecordNotFoundError: ignoreRecordNotFound,
		LogLevel:                  gormlogger.Info,
	}
}

// LogMode implementation of gormlogger.Interface
func (l *ZapGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info implementation of gormlogger.Interface
func (l *ZapGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		log.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn implementation of gormlogger.Interface
func (l *ZapGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		log.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error implementation of gormlogger.Interface
func (l *ZapGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		log.Error(fmt.Sprintf(msg, data...))
	}
}

// Trace implementation of gormlogger.Interface
func (l *ZapGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)

	// Skip log for record not found error
	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && l.IgnoreRecordNotFoundError) {
		sql, rows := fc()
		log.Error("SQL Error",
			zap.Error(err),
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("elapsed", elapsed),
		)
		return
	}

	// Log slow queries
	if l.SlowThreshold > 0 && elapsed > l.SlowThreshold && l.LogLevel >= gormlogger.Warn {
		sql, rows := fc()
		log.Warn("Slow SQL query",
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("elapsed", elapsed),
			zap.Duration("threshold", l.SlowThreshold),
		)
	}
}

// Connect establishes connection to the database
func Connect(cfg *config.Config) (*gorm.DB, error) {
	// Initialize the logger before using it
	log.Logger() // Ensure logger is initialized

	// Configure GORM logger
	gormLogger := NewZapGormLogger(time.Second, true)

	// Set log level based on environment
	logLevel := gormlogger.Info
	if cfg.Environment == "production" {
		logLevel = gormlogger.Error
	} else if cfg.Environment == "development" {
		logLevel = gormlogger.Info
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: gormLogger.LogMode(logLevel),
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

	log.Info("Connected to database")
	return db, nil
}

// MigrateSchema performs database migrations
func MigrateSchema(db *gorm.DB) error {
	log.Info("Running database migrations...")

	// Auto-migrate all models
	err := db.AutoMigrate(
		&models.User{},
		&models.Theme{},
		&models.Board{},
		&models.BoardContributor{},
		&models.Post{},
		&models.PostLike{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Info("Database migrations completed")
	return nil
}
