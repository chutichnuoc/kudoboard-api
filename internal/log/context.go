package log

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ContextLogger creates a logger with context information from a Gin context
func ContextLogger(c *gin.Context) *zap.Logger {
	// Start with the global logger
	logger := Logger()

	// Add request ID if available
	if requestID, exists := c.Get("RequestID"); exists {
		logger = logger.With(zap.String("request_id", requestID.(string)))
	}

	// Add user ID if available
	if userID, exists := c.Get("userID"); exists && userID != uint(0) {
		logger = logger.With(zap.Uint("user_id", userID.(uint)))
	}

	// Add IP address
	logger = logger.With(zap.String("ip", c.ClientIP()))

	// Add path and method
	logger = logger.With(
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	)

	return logger.WithOptions(zap.AddCallerSkip(-1), zap.AddStacktrace(zap.FatalLevel))
}

// ContextSugar creates a sugared logger with context information
func ContextSugar(c *gin.Context) *zap.SugaredLogger {
	return ContextLogger(c).Sugar()
}
