package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"strings"

	"telegram-bot/internal/domain"
	"telegram-bot/internal/usecase"
)

// LoggingMiddleware creates request logging middleware
func LoggingMiddleware(requestLogUsecase *usecase.RequestLogUsecase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Record start time
		startTime := time.Now()

		// Process request
		err := c.Next()

		// Calculate response time
		responseTime := time.Since(startTime).Milliseconds()

		// Get user ID from context if available
		var userID *uuid.UUID
		if userIDValue := c.Locals("user_id"); userIDValue != nil {
			if uid, ok := userIDValue.(uuid.UUID); ok {
				userID = &uid
			}
		}

		// Create request log
		requestLog := &domain.CreateRequestLogRequest{
			IPAddress:    c.IP(),
			HTTPMethod:   c.Method(),
			Path:         c.Path(),
			UserAgent:    c.Get("User-Agent"),
			StatusCode:   c.Response().StatusCode(),
			ResponseTime: responseTime,
			UserID:       userID,
		}

		// Log request asynchronously to avoid blocking the response
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if logErr := requestLogUsecase.LogRequest(ctx, requestLog); logErr != nil {
				// In production, you might want to use a proper logger here
				// For now, we'll silently ignore logging errors to not affect the main request
				_ = logErr
			}
		}()

		return err
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Generate unique request ID
		requestID := uuid.New().String()

		// Set request ID in context and response header
		c.Locals("request_id", requestID)
		c.Set("X-Request-ID", requestID)

		return c.Next()
	}
}

// CORSMiddleware handles CORS headers
func CORSMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Set CORS headers
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Set("Access-Control-Expose-Headers", "X-Request-ID")

		// Handle preflight requests
		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}

// SecurityMiddleware adds security headers
func SecurityMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Security headers
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Set CSP with allowances for Swagger
		path := c.Path()
		if strings.HasPrefix(path, "/swagger/") {
			c.Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; script-src 'self' 'unsafe-inline'")
		} else {
			c.Set("Content-Security-Policy", "default-src 'self'")
		}

		return c.Next()
	}
}