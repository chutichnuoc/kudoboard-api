package middleware

import (
	"bytes"
	"io"
	"kudoboard-api/internal/log"
	"kudoboard-api/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// responseWriter is a wrapper around http.ResponseWriter to capture response data
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the response body while writing it
func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LoggingMiddleware adds structured log to all requests
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Get request data
		path := c.Request.URL.Path
		method := c.Request.Method
		query := c.Request.URL.RawQuery
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Generate a request ID if not already set
		var requestID string
		if id, exists := c.Get("RequestID"); exists {
			requestID = id.(string)
		}

		// Log the request
		log.Info("Request started",
			zap.String("path", path),
			zap.String("method", method),
			zap.String("query", query),
			zap.String("ip", clientIP),
			zap.String("user-agent", userAgent),
			zap.String("request_id", requestID),
		)

		// Parse and log request body if needed (not for multipart/form-data to avoid file upload issues)
		if c.Request.Method != http.MethodGet && c.ContentType() != "multipart/form-data" {
			if body, err := io.ReadAll(c.Request.Body); err == nil {
				// Reset the request body for future middleware
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
				// Only log body for non-binary content
				if len(body) > 0 && utils.IsTextContent(c.ContentType()) {
					log.Debug("Request body",
						zap.String("body", string(body)),
						zap.String("request_id", requestID),
					)
				}
			}
		}

		// Create a custom response writer to capture the response
		respWriter := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = respWriter

		// Process request
		c.Next()

		// Get response status and latency
		status := c.Writer.Status()
		latency := time.Since(start)

		// Create log fields
		fields := []zap.Field{
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.Int("size", c.Writer.Size()),
			zap.String("request_id", requestID),
		}

		// Add user ID if available
		if userID, exists := c.Get("userID"); exists && userID.(uint) != 0 {
			fields = append(fields, zap.Uint("user_id", userID.(uint)))
		}

		// Add response body for error responses
		if status >= 400 && utils.IsTextContent(c.Writer.Header().Get("Content-Type")) {
			respBody := respWriter.body.String()
			if len(respBody) > 0 {
				fields = append(fields, zap.String("response", respBody))
			}
		}

		// Log based on status code - but don't log errors that will be handled by error middleware
		logMsg := "Request completed"
		if status >= 500 {
			// Only log server errors here if they weren't already handled by error middleware
			if len(c.Errors) == 0 {
				log.Error(logMsg, fields...)
			}
		} else if status >= 400 {
			// Only log client errors here if they weren't already handled by error middleware
			if len(c.Errors) == 0 {
				log.Warn(logMsg, fields...)
			}
		} else {
			log.Info(logMsg, fields...)
		}
	}
}

// RequestIDMiddleware adds a unique request ID to each request context
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := utils.GenerateRequestID()
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}
