package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"telegram-bot/pkg/jwt"
	"telegram-bot/pkg/response"
)

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware(jwtManager *jwt.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Unauthorized(c, "Authorization header is required")
		}

		// Check Bearer token format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return response.Unauthorized(c, "Invalid authorization header format")
		}

		tokenString := tokenParts[1]

		// Validate token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			return response.Unauthorized(c, "Invalid or expired token")
		}

		// Check if it's a temporary token (should not be used for protected routes)
		if claims.TempAuth {
			return response.Unauthorized(c, "Temporary token cannot be used for this endpoint")
		}

		// Store user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("is_2fa", claims.Is2FA)

		return c.Next()
	}
}

// OptionalAuthMiddleware creates optional JWT authentication middleware
// This middleware will extract user info if token is present but won't fail if it's not
func OptionalAuthMiddleware(jwtManager *jwt.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		// Check Bearer token format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return c.Next()
		}

		tokenString := tokenParts[1]

		// Validate token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			return c.Next()
		}

		// Skip temporary tokens
		if claims.TempAuth {
			return c.Next()
		}

		// Store user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("is_2fa", claims.Is2FA)

		return c.Next()
	}
}

// TempAuthMiddleware creates middleware for temporary token validation (for 2FA)
func TempAuthMiddleware(jwtManager *jwt.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Unauthorized(c, "Authorization header is required")
		}

		// Check Bearer token format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return response.Unauthorized(c, "Invalid authorization header format")
		}

		tokenString := tokenParts[1]

		// Validate temporary token
		claims, err := jwtManager.ValidateTempToken(tokenString)
		if err != nil {
			return response.Unauthorized(c, "Invalid or expired temporary token")
		}

		// Store user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("temp_token", tokenString)

		return c.Next()
	}
}