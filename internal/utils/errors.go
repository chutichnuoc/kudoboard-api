package utils

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// Custom error types
var (
	ErrNotFound      = errors.New("resource not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrBadRequest    = errors.New("bad request")
	ErrInternalError = errors.New("internal server error")
	ErrValidation    = errors.New("validation error")
)

// AppError represents an application error with additional context
type AppError struct {
	Code        string                 // Error code for client
	Message     string                 // User-friendly message
	Err         error                  // Original error
	stack       string                 // Stack trace
	OperationID string                 // Optional operation ID for tracking
	Fields      map[string]interface{} // Additional context fields
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithField adds a context field to the error
func (e *AppError) WithField(key string, value interface{}) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]interface{})
	}
	e.Fields[key] = value
	return e
}

// WithFields adds multiple context fields to the error
func (e *AppError) WithFields(fields map[string]interface{}) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]interface{})
	}
	for k, v := range fields {
		e.Fields[k] = v
	}
	return e
}

// WithOperationID adds an operation ID for tracking
func (e *AppError) WithOperationID(id string) *AppError {
	e.OperationID = id
	return e
}

// CaptureStack captures the current stack trace
func (e *AppError) CaptureStack() *AppError {
	const depth = 32
	var pcs [depth]uintptr
	// Skip 3 frames to avoid including the error creation machinery
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var sb strings.Builder
	for {
		frame, more := frames.Next()
		// Skip framework code
		if strings.Contains(frame.File, "gin-gonic") ||
			strings.Contains(frame.Function, "kudoboard-api/internal/utils.") {
			if !more {
				break
			}
			continue
		}

		sb.WriteString(fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}

	e.stack = sb.String()
	return e
}

// GetStack returns the captured stack trace
func (e *AppError) GetStack() string {
	return e.stack
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError converts an error to an AppError if it isn't already
func AsAppError(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	// Use a default error code based on the error type
	code := "INTERNAL_ERROR"
	switch {
	case errors.Is(err, ErrNotFound):
		code = "NOT_FOUND"
	case errors.Is(err, ErrBadRequest):
		code = "BAD_REQUEST"
	case errors.Is(err, ErrUnauthorized):
		code = "UNAUTHORIZED"
	case errors.Is(err, ErrForbidden):
		code = "FORBIDDEN"
	case errors.Is(err, ErrValidation):
		code = "VALIDATION_ERROR"
	}

	return &AppError{
		Code:    code,
		Message: err.Error(),
		Err:     err,
	}
}

// Error creation helpers

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Code:    "NOT_FOUND",
		Message: message,
		Err:     ErrNotFound,
	}
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:    "UNAUTHORIZED",
		Message: message,
		Err:     ErrUnauthorized,
	}
}

// NewForbiddenError creates a new forbidden error
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:    "FORBIDDEN",
		Message: message,
		Err:     ErrForbidden,
	}
}

// NewBadRequestError creates a new bad request error
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:    "BAD_REQUEST",
		Message: message,
		Err:     ErrBadRequest,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *AppError {
	return &AppError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Err:     ErrValidation,
	}
}

// NewInternalError creates a new internal server error
func NewInternalError(message string, err error) *AppError {
	appErrpr := &AppError{
		Code:    "INTERNAL_ERROR",
		Message: message,
		Err:     errors.Join(ErrInternalError, err),
	}
	_ = appErrpr.CaptureStack()
	return appErrpr
}

// WrapError wraps an existing error with additional context
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		// Create a copy to avoid modifying the original
		newAppErr := *appErr
		newAppErr.Message = message + ": " + appErr.Message
		return &newAppErr
	}

	return fmt.Errorf("%s: %w", message, err)
}
