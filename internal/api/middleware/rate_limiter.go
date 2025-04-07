package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/dto/responses"
	"kudoboard-api/internal/log"
)

// Client represents a client with its rate limiter
type Client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiterMiddleware contains rate limiting logic
type RateLimiterMiddleware struct {
	clients     map[string]*Client
	mu          sync.Mutex
	cfg         *config.Config
	cleanup     *time.Ticker
	done        chan bool
	ipLimiter   *rate.Limiter // Global IP-based limiter
	authLimiter *rate.Limiter // Auth endpoint specific limiter
}

// NewRateLimiterMiddleware creates a new rate limiter middleware
func NewRateLimiterMiddleware(cfg *config.Config) *RateLimiterMiddleware {
	r := &RateLimiterMiddleware{
		clients:     make(map[string]*Client),
		cfg:         cfg,
		cleanup:     time.NewTicker(time.Minute * 5), // Clean up unused clients every 5 minutes
		done:        make(chan bool),
		ipLimiter:   rate.NewLimiter(rate.Limit(cfg.RateLimitRequests), cfg.RateLimitBurst),         // Default IP limiter
		authLimiter: rate.NewLimiter(rate.Limit(cfg.AuthRateLimitRequests), cfg.AuthRateLimitBurst), // Auth specific limiter
	}

	// Start cleanup goroutine
	go r.cleanupClients()

	return r
}

// getClientLimiter gets or creates a limiter for a client
func (r *RateLimiterMiddleware) getClientLimiter(key string, isAuth bool) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	client, exists := r.clients[key]
	if !exists {
		var limiter *rate.Limiter
		if isAuth {
			limiter = rate.NewLimiter(rate.Limit(r.cfg.AuthRateLimitRequests), r.cfg.AuthRateLimitBurst)
		} else {
			limiter = rate.NewLimiter(rate.Limit(r.cfg.RateLimitRequests), r.cfg.RateLimitBurst)
		}
		client = &Client{limiter: limiter, lastSeen: time.Now()}
		r.clients[key] = client
		return limiter
	}

	// Update last seen time
	client.lastSeen = time.Now()
	return client.limiter
}

// cleanupClients removes clients that haven't been seen for a while
func (r *RateLimiterMiddleware) cleanupClients() {
	for {
		select {
		case <-r.cleanup.C:
			r.mu.Lock()
			for ip, client := range r.clients {
				if time.Since(client.lastSeen) > time.Hour {
					delete(r.clients, ip)
				}
			}
			r.mu.Unlock()
		case <-r.done:
			r.cleanup.Stop()
			return
		}
	}
}

// Shutdown stops the cleanup goroutine
func (r *RateLimiterMiddleware) Shutdown() {
	close(r.done)
}

// RateLimit creates a gin middleware for rate limiting
func (r *RateLimiterMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		// Check if this is an auth endpoint
		isAuth := c.Request.URL.Path == "/api/v1/auth/login" ||
			c.Request.URL.Path == "/api/v1/auth/register" ||
			c.Request.URL.Path == "/api/v1/auth/google" ||
			c.Request.URL.Path == "/api/v1/auth/facebook" ||
			c.Request.URL.Path == "/api/v1/auth/forgot-password" ||
			c.Request.URL.Path == "/api/v1/auth/reset-password"

		// Get the appropriate limiter
		limiter := r.getClientLimiter(clientIP, isAuth)

		// Check if allowed
		if !limiter.Allow() {
			log.Warn("Rate limit exceeded",
				zap.String("ip", clientIP),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)

			c.JSON(http.StatusTooManyRequests, responses.ErrorResponse(
				"RATE_LIMIT_EXCEEDED",
				"You have exceeded the request rate limit. Please try again later.",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
