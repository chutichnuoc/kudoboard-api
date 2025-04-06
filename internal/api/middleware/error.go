package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"kudoboard-api/internal/dto/responses"
	"kudoboard-api/internal/log"
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
				// Log the panic with stack trace
				stackTrace := string(debug.Stack())
				logger := log.ContextLogger(c)
				logger.Error("Panic recovered",
					zap.Any("panic", r),
					zap.String("stack", stackTrace),
				)

				// Create response with appropriate detail level
				errorMessage := "An unexpected error occurred"
				if m.Debug {
					errorMessage = fmt.Sprintf("%v", r)
				}

				// Only write response if it hasn't been written yet
				if !c.Writer.Written() {
					appError := utils.NewInternalError(errorMessage, fmt.Errorf("%v", r))

					c.JSON(http.StatusInternalServerError, responses.ErrorResponse(
						appError.Code,
						errorMessage,
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

			// If response hasn't been written yet
			if !c.Writer.Written() {
				statusCode, errorResponse := m.processError(err, c)
				c.JSON(statusCode, errorResponse)
			}
		}
	}
}

// processError analyzes the error and returns appropriate status code and response
func (m *ErrorMiddleware) processError(err error, c *gin.Context) (int, responses.APIResponse) {
	logger := log.ContextLogger(c)

	// Check if it's our app error type
	var appError *utils.AppError
	if !errors.As(err, &appError) {
		// Convert to app error
		appError = utils.AsAppError(err)
	}

	// Log the error with contextual information
	logFields := []zap.Field{
		zap.String("error_code", appError.Code),
		zap.Error(err),
	}

	if appError.Fields != nil {
		for key, value := range appError.Fields {
			logFields = append(logFields, zap.Any(key, value))
		}
	}

	if m.Debug && appError.GetStack() != "" {
		logFields = append(logFields, zap.String("stack", appError.GetStack()))
	}

	// Log based on error type
	switch {
	case errors.Is(err, utils.ErrNotFound):
		logger.Info("Resource not found", logFields...)
	case errors.Is(err, utils.ErrBadRequest) || errors.Is(err, utils.ErrValidation):
		logger.Info("Bad request", logFields...)
	case errors.Is(err, utils.ErrUnauthorized):
		logger.Info("Unauthorized access attempt", logFields...)
	case errors.Is(err, utils.ErrForbidden):
		logger.Warn("Forbidden access attempt", logFields...)
	default:
		logger.Error("Internal server error", logFields...)
	}

	// Map the error to HTTP status code and create response
	statusCode := m.mapErrorToStatusCode(appError)
	response := responses.ErrorResponse(
		appError.Code,
		appError.Message,
	)

	// Add details if in debug mode
	if m.Debug {
		details := m.buildErrorDetails(appError)
		if details != "" {
			response.Error.Details = details
		}
	}

	return statusCode, response
}

// mapErrorToStatusCode maps app error to HTTP status code
func (m *ErrorMiddleware) mapErrorToStatusCode(appError *utils.AppError) int {
	switch {
	case errors.Is(appError.Err, utils.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(appError.Err, utils.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(appError.Err, utils.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(appError.Err, utils.ErrBadRequest):
		return http.StatusBadRequest
	case errors.Is(appError.Err, utils.ErrValidation):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// buildErrorDetails creates detailed error information for debug mode
func (m *ErrorMiddleware) buildErrorDetails(appError *utils.AppError) string {
	var details []string

	// Add original error
	if appError.Err != nil && !errors.Is(appError.Err, utils.ErrInternalError) &&
		!errors.Is(appError.Err, utils.ErrBadRequest) && !errors.Is(appError.Err, utils.ErrNotFound) &&
		!errors.Is(appError.Err, utils.ErrForbidden) && !errors.Is(appError.Err, utils.ErrUnauthorized) {
		details = append(details, fmt.Sprintf("Cause: %v", appError.Err))
	}

	// Add stack trace
	if stack := appError.GetStack(); stack != "" {
		details = append(details, fmt.Sprintf("Stack: %s", stack))
	}

	// Add operation ID if present
	if appError.OperationID != "" {
		details = append(details, fmt.Sprintf("Operation: %s", appError.OperationID))
	}

	return strings.Join(details, "\n")
}

// NotFoundHandler handles 404 errors
func (m *ErrorMiddleware) NotFoundHandler(c *gin.Context) {
	log.ContextLogger(c).Info("Resource not found",
		zap.String("path", c.Request.URL.Path),
	)

	c.JSON(http.StatusNotFound, responses.ErrorResponse(
		"NOT_FOUND",
		fmt.Sprintf("The requested resource '%s' could not be found", c.Request.URL.Path),
	))
}

// MethodNotAllowedHandler handles 405 errors
func (m *ErrorMiddleware) MethodNotAllowedHandler(c *gin.Context) {
	log.ContextLogger(c).Info("Method not allowed",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
	)

	c.JSON(http.StatusMethodNotAllowed, responses.ErrorResponse(
		"METHOD_NOT_ALLOWED",
		fmt.Sprintf("Method '%s' is not allowed for this resource", c.Request.Method),
	))
}
