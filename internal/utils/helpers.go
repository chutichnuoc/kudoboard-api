package utils

import (
	"github.com/google/uuid"
	"strings"
)

// GenerateRequestID creates a unique identifier for each request
func GenerateRequestID() string {
	return uuid.New().String()
}

// IsTextContent checks if a content type is text-based for safe log
func IsTextContent(contentType string) bool {
	if contentType == "" {
		return false
	}

	textTypes := []string{
		"application/json",
		"application/xml",
		"application/javascript",
		"text/",
	}

	for _, textType := range textTypes {
		if strings.Contains(contentType, textType) {
			return true
		}
	}

	return false
}

// FormatLogError formats an error for log purposes
func FormatLogError(err error) string {
	if err == nil {
		return ""
	}

	// If it's custom app error, get underlying error
	if appErr, ok := err.(*AppError); ok {
		if appErr.Err != nil {
			return appErr.Err.Error()
		}
		return appErr.Message
	}

	return err.Error()
}
