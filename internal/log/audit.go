package log

import (
	"time"

	"go.uber.org/zap"
)

// AuditLog represents a security-related log entry for sensitive operations
type AuditLog struct {
	Action     string    // The action performed (e.g., "login", "update_user", "delete_board")
	UserID     uint      // The ID of the user performing the action
	TargetType string    // The type of resource being acted upon (e.g., "user", "board", "post")
	TargetID   uint      // The ID of the resource being acted upon
	Details    string    // Additional details about the action
	Status     string    // Result status (e.g., "success", "failure")
	IP         string    // IP address of the requester
	RequestID  string    // Unique request ID for correlation
	Timestamp  time.Time // When the action occurred
}

// LogAudit logs an audit event for security tracking
func LogAudit(log AuditLog) {
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	Logger().Info("Audit event",
		zap.String("action", log.Action),
		zap.Uint("user_id", log.UserID),
		zap.String("target_type", log.TargetType),
		zap.Uint("target_id", log.TargetID),
		zap.String("details", log.Details),
		zap.String("status", log.Status),
		zap.String("ip", log.IP),
		zap.String("request_id", log.RequestID),
		zap.Time("timestamp", log.Timestamp),
	)
}

// LogAuthAttempt logs authentication attempts (success or failure)
func LogAuthAttempt(userID uint, email string, status string, ip string, requestID string, details string) {
	Logger().Info("Authentication attempt",
		zap.Uint("user_id", userID),
		zap.String("email", email),
		zap.String("status", status),
		zap.String("ip", ip),
		zap.String("request_id", requestID),
		zap.String("details", details),
		zap.Time("timestamp", time.Now()),
	)
}

// LogResourceAccess logs access to sensitive resources
func LogResourceAccess(userID uint, resourceType string, resourceID uint, action string, ip string, requestID string) {
	Logger().Info("Resource access",
		zap.Uint("user_id", userID),
		zap.String("resource_type", resourceType),
		zap.Uint("resource_id", resourceID),
		zap.String("action", action),
		zap.String("ip", ip),
		zap.String("request_id", requestID),
		zap.Time("timestamp", time.Now()),
	)
}

// LogSecurity logs security-related events
func LogSecurity(event string, userID uint, ip string, requestID string, details string) {
	Logger().Warn("Security event",
		zap.String("event", event),
		zap.Uint("user_id", userID),
		zap.String("ip", ip),
		zap.String("request_id", requestID),
		zap.String("details", details),
		zap.Time("timestamp", time.Now()),
	)
}
