package middleware

import (
	"github.com/gin-gonic/gin"
	"kudoboard-api/internal/dto/responses"
	"kudoboard-api/internal/log"
	"kudoboard-api/internal/models"
	"net/http"

	"go.uber.org/zap"
)

// AdminOnly ensures that only admin users can access the route
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request ID for correlation
		requestID, _ := c.Get("RequestID")
		requestIDStr, _ := requestID.(string)

		// Get user from context (set by auth middleware)
		user, exists := c.Get("user")
		if !exists {
			log.Warn("Admin access attempt without authentication",
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()),
				zap.String("request_id", requestIDStr),
			)

			c.JSON(http.StatusUnauthorized, responses.ErrorResponse("UNAUTHORIZED", "Authentication required"))
			c.Abort()
			return
		}

		typedUser, ok := user.(*models.User)
		if !ok || !isAdmin(typedUser) {
			// Log the unauthorized admin access attempt
			log.Warn("Unauthorized admin access attempt",
				zap.Uint("user_id", typedUser.ID),
				zap.String("email", typedUser.Email),
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()),
				zap.String("request_id", requestIDStr),
			)

			c.JSON(http.StatusForbidden, responses.ErrorResponse("FORBIDDEN", "Admin access required"))
			c.Abort()
			return
		}

		// Log successful admin access
		log.Info("Admin access granted",
			zap.Uint("user_id", typedUser.ID),
			zap.String("email", typedUser.Email),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.String("request_id", requestIDStr),
		)

		c.Next()
	}
}

// isAdmin checks if a user has admin privileges
func isAdmin(user *models.User) bool {
	return user.IsAdmin
}
