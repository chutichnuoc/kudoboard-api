package middleware

import (
	"github.com/gin-gonic/gin"
	"kudoboard-api/internal/dto/responses"
	"kudoboard-api/internal/models"
	"net/http"
)

// AdminOnly ensures that only admin users can access the route
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (set by auth middleware)
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, responses.ErrorResponse("UNAUTHORIZED", "Authentication required"))
			c.Abort()
			return
		}

		typedUser, ok := user.(*models.User)
		if !ok || !isAdmin(typedUser) {
			c.JSON(http.StatusForbidden, responses.ErrorResponse("FORBIDDEN", "Admin access required"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// isAdmin checks if a user has admin privileges
func isAdmin(user *models.User) bool {
	return user.IsAdmin
}
