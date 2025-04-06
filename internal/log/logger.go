package log

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Global logger instance
	logger *zap.Logger
	// Global sugared logger instance for convenience
	sugar *zap.SugaredLogger
	once  sync.Once
)

// Logger returns the global logger instance
func Logger() *zap.Logger {
	once.Do(initLogger)
	return logger
}

// Sugar returns the global sugared logger instance
func Sugar() *zap.SugaredLogger {
	once.Do(initLogger)
	return sugar
}

// initLogger initializes the logger with appropriate configuration
func initLogger() {
	// Default to development mode
	environment := os.Getenv("APP_ENV")
	if environment == "" {
		environment = "development"
	}

	var config zap.Config
	if environment == "production" {
		// Production config: JSON format, info level
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		// Development config: console format, debug level
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var err error
	logger, err = config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.FatalLevel),
	)
	if err != nil {
		// Fallback to a basic logger if setup fails
		logger = zap.NewExample()
		logger.Error("Failed to initialize logger", zap.Error(err))
	}

	sugar = logger.Sugar()
}

// Debug logs a message at debug level
func Debug(msg string, fields ...zapcore.Field) {
	Logger().Debug(msg, fields...)
}

// Info logs a message at info level
func Info(msg string, fields ...zapcore.Field) {
	Logger().Info(msg, fields...)
}

// Warn logs a message at warn level
func Warn(msg string, fields ...zapcore.Field) {
	Logger().Warn(msg, fields...)
}

// Error logs a message at error level
func Error(msg string, fields ...zapcore.Field) {
	Logger().Error(msg, fields...)
}

// Fatal logs a message at fatal level and then exits
func Fatal(msg string, fields ...zapcore.Field) {
	Logger().Fatal(msg, fields...)
}

// With returns a Logger with the specified fields added
func With(fields ...zapcore.Field) *zap.Logger {
	return Logger().With(fields...)
}

// Named returns a Logger with the specified name
func Named(name string) *zap.Logger {
	return Logger().Named(name)
}

// Sync flushes any buffered log entries
func Sync() error {
	return Logger().Sync()
}

// Shutdown gracefully shuts down the logger, ensuring all logs are flushed
func Shutdown() {
	_ = Logger().Sync()
}
