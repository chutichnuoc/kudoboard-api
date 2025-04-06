package middleware

import (
	"github.com/gin-gonic/gin"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/dto/responses"
	"kudoboard-api/internal/log"
	"kudoboard-api/internal/services"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// AuthMiddleware handles authentication for protected routes
type AuthMiddleware struct {
	authService *services.AuthService
	cfg         *config.Config
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authService *services.AuthService, cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		cfg:         cfg,
	}
}

// RequireAuth creates a middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the authorization header
		authHeader := c.GetHeader("Authorization")

		// Get request ID for correlation
		requestID, _ := c.Get("RequestID")
		requestIDStr, _ := requestID.(string)

		// Check if auth header exists and has the correct format
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			log.Info("Authentication failed: missing or invalid token format",
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()),
				zap.String("request_id", requestIDStr),
			)

			c.JSON(http.StatusUnauthorized, responses.ErrorResponse("UNAUTHORIZED", "Authentication required"))
			c.Abort()
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Verify token using auth service
		userID, err := m.authService.VerifyToken(tokenString)
		if err != nil {
			log.Info("Authentication failed: invalid token",
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()),
				zap.String("error", err.Error()),
				zap.String("request_id", requestIDStr),
			)

			c.JSON(http.StatusUnauthorized, responses.ErrorResponse("INVALID_TOKEN", "Invalid or expired token"))
			c.Abort()
			return
		}

		// Get the user
		user, err := m.authService.GetUserByID(userID)
		if err != nil {
			log.Info("Authentication failed: user not found",
				zap.Uint("user_id", userID),
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()),
				zap.String("request_id", requestIDStr),
			)

			c.JSON(http.StatusUnauthorized, responses.ErrorResponse("USER_NOT_FOUND", "User not found"))
			c.Abort()
			return
		}

		// Set the user and userID in the context
		c.Set("user", user)
		c.Set("userID", userID)

		log.Info("User authenticated",
			zap.Uint("user_id", userID),
			zap.String("path", c.Request.URL.Path),
			zap.String("request_id", requestIDStr),
		)

		c.Next()
	}
}

// OptionalAuth creates a middleware that attempts authentication but doesn't require it
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the authorization header
		authHeader := c.GetHeader("Authorization")

		// Get request ID for correlation
		requestID, _ := c.Get("RequestID")
		requestIDStr, _ := requestID.(string)

		// If no auth header, continue without user
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Try to verify token
		userID, err := m.authService.VerifyToken(tokenString)
		if err != nil {
			// Log but don't abort
			log.Debug("Optional auth: invalid token",
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()),
				zap.String("error", err.Error()),
				zap.String("request_id", requestIDStr),
			)

			// Don't abort, just continue without user
			c.Next()
			return
		}

		// Try to get the user
		user, err := m.authService.GetUserByID(userID)
		if err != nil {
			// Log but don't abort
			log.Debug("Optional auth: user not found",
				zap.Uint("user_id", userID),
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()),
				zap.String("request_id", requestIDStr),
			)

			// Don't abort, just continue without user
			c.Next()
			return
		}

		// Set the user and userID in the context
		c.Set("user", user)
		c.Set("userID", userID)

		log.Debug("Optional auth: user authenticated",
			zap.Uint("user_id", userID),
			zap.String("path", c.Request.URL.Path),
			zap.String("request_id", requestIDStr),
		)

		c.Next()
	}
}
