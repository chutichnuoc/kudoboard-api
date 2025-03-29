package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"kudoboard-api/internal/dto/responses"
	"kudoboard-api/internal/utils"
)

// ErrorMiddleware handles errors globally
type ErrorMiddleware struct {
	Debug bool // Enable detailed error information in development
}

// NewErrorMiddleware creates a new ErrorMiddleware instance
func NewErrorMiddleware(debug bool) *ErrorMiddleware {
	return &ErrorMiddleware{
		Debug: debug,
	}
}

// ErrorHandler combines both panic recovery and error handling in a single middleware
func (m *ErrorMiddleware) ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set up recovery from panics
		defer func() {
			if r := recover(); r != nil {
				// Log the panic
				stackTrace := string(debug.Stack())
				log.Printf("Panic recovered: %v\nStack Trace:\n%s", r, stackTrace)

				// Determine if we should show stack trace in the response
				var errorDetails string
				if m.Debug {
					errorDetails = fmt.Sprintf("%v", r)
				} else {
					errorDetails = "An unexpected error occurred"
				}

				// If response is already written, don't attempt to write again
				if !c.Writer.Written() {
					// Return a 500 response
					c.JSON(http.StatusInternalServerError, responses.ErrorResponse(
						"INTERNAL_SERVER_ERROR",
						errorDetails,
					))
				}
				c.Abort()
			}
		}()

		// Process the request
		c.Next()

		// Handle any errors added to the context
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Log the error
			if m.Debug {
				log.Printf("Error: %v\nStack Trace:\n%s", err, debug.Stack())
			} else {
				log.Printf("Error: %v", err)
			}

			// If response hasn't been written yet (by panic recovery or other middleware)
			if !c.Writer.Written() {
				statusCode, errorResponse := m.processError(err, c.Request.Method, c.Request.URL.Path)
				c.JSON(statusCode, errorResponse)
			}
		}
	}
}

// processError analyzes the error and returns appropriate status code and response
func (m *ErrorMiddleware) processError(err error, method, path string) (int, responses.APIResponse) {
	// Check if it's one of our app errors
	var appError *utils.AppError
	if errors.As(err, &appError) {
		return m.handleAppError(appError)
	}

	// Handle validation errors (from gin binding)
	if strings.Contains(err.Error(), "binding") || strings.Contains(err.Error(), "validate") {
		return http.StatusBadRequest, responses.ErrorResponse(
			"VALIDATION_ERROR",
			err.Error(),
		)
	}

	// Handle database errors
	if strings.Contains(err.Error(), "database") || strings.Contains(err.Error(), "sql") {
		// Don't expose detailed database errors to clients
		errorMsg := "Database operation failed"
		if m.Debug {
			errorMsg = err.Error()
		}
		return http.StatusInternalServerError, responses.ErrorResponse(
			"DATABASE_ERROR",
			errorMsg,
		)
	}

	// Handle authentication errors
	if strings.Contains(err.Error(), "token") || strings.Contains(err.Error(), "auth") {
		return http.StatusUnauthorized, responses.ErrorResponse(
			"AUTHENTICATION_ERROR",
			err.Error(),
		)
	}

	// Default to internal server error for unhandled error types
	errorMsg := "An unexpected error occurred"
	if m.Debug {
		errorMsg = err.Error()
	}

	// Log the unhandled error for investigation
	log.Printf("Unhandled error type on %s %s: %v", method, path, err)

	return http.StatusInternalServerError, responses.ErrorResponse(
		"INTERNAL_SERVER_ERROR",
		errorMsg,
	)
}

// handleAppError processes application-specific errors
func (m *ErrorMiddleware) handleAppError(appError *utils.AppError) (int, responses.APIResponse) {
	// Map app error to HTTP status code
	var statusCode int

	switch {
	case errors.Is(appError.Err, utils.ErrNotFound):
		statusCode = http.StatusNotFound
	case errors.Is(appError.Err, utils.ErrUnauthorized):
		statusCode = http.StatusUnauthorized
	case errors.Is(appError.Err, utils.ErrForbidden):
		statusCode = http.StatusForbidden
	case errors.Is(appError.Err, utils.ErrBadRequest):
		statusCode = http.StatusBadRequest
	case errors.Is(appError.Err, utils.ErrInternalError):
		statusCode = http.StatusInternalServerError
	default:
		statusCode = http.StatusInternalServerError
	}

	// Create response
	response := responses.ErrorResponse(
		appError.Code,
		appError.Message,
	)

	// Add details if in debug mode and we have an underlying error
	if m.Debug && errors.Unwrap(appError) != nil {
		detailErr := errors.Unwrap(appError)
		response.Error.Details = detailErr.Error()
	}

	return statusCode, response
}

// NotFoundHandler handles 404 errors
func (m *ErrorMiddleware) NotFoundHandler(c *gin.Context) {
	log.Printf("Not found: %s %s", c.Request.Method, c.Request.URL.Path)

	c.JSON(http.StatusNotFound, responses.ErrorResponse(
		"NOT_FOUND",
		fmt.Sprintf("The requested resource '%s' could not be found", c.Request.URL.Path),
	))
}

// MethodNotAllowedHandler handles 405 errors
func (m *ErrorMiddleware) MethodNotAllowedHandler(c *gin.Context) {
	log.Printf("Method not allowed: %s %s", c.Request.Method, c.Request.URL.Path)

	c.JSON(http.StatusMethodNotAllowed, responses.ErrorResponse(
		"METHOD_NOT_ALLOWED",
		fmt.Sprintf("Method '%s' is not allowed for this resource", c.Request.Method),
	))
}
