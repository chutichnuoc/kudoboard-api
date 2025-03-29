package middleware

import (
	"github.com/gin-gonic/gin"
	"kudoboard-api/internal/dto/responses"
	"net/http"
)

// ErrorMiddleware handles errors globally
type ErrorMiddleware struct{}

// ErrorHandler is a middleware that handles errors globally
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Execute request handlers
		c.Next()

		// If there are errors, handle them
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Check error type and return appropriate response
			// This is a simplified version - in a real application, you'd handle different error types
			c.JSON(http.StatusInternalServerError, responses.ErrorResponse(
				"INTERNAL_SERVER_ERROR",
				err.Error(),
			))
		}
	}
}

// NotFoundHandler handles 404 errors
func NotFoundHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, responses.ErrorResponse(
		"NOT_FOUND",
		"The requested resource could not be found",
	))
}
