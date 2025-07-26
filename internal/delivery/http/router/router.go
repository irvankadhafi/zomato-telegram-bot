package router

import (
	"github.com/gofiber/fiber/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	_ "telegram-bot/docs" // Import docs for swagger
	"telegram-bot/internal/delivery/http/handler"
	"telegram-bot/internal/delivery/http/middleware"
	"telegram-bot/internal/usecase"
	"telegram-bot/pkg/jwt"
)

// RouterConfig holds the configuration for setting up routes
type RouterConfig struct {
	App                *fiber.App
	AuthUsecase        *usecase.AuthUsecase
	RestaurantUsecase  *usecase.RestaurantUsecase
	RequestLogUsecase  *usecase.RequestLogUsecase
	JWTManager         *jwt.JWTManager
	TelegramBot        *tgbotapi.BotAPI
}

// SetupRoutes sets up all application routes
func SetupRoutes(config *RouterConfig) {
	app := config.App

	// Create handlers
	authHandler := handler.NewAuthHandler(config.AuthUsecase)
	restaurantHandler := handler.NewRestaurantHandler(config.RestaurantUsecase)
	telegramHandler := handler.NewTelegramHandler(config.TelegramBot, config.RestaurantUsecase)

	// Global middleware
	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.CORSMiddleware())
	app.Use(middleware.SecurityMiddleware())
	app.Use(middleware.LoggingMiddleware(config.RequestLogUsecase))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "telegram-bot-api",
			"version": "1.0.0",
		})
	})

	// API v1 routes
	v1 := app.Group("/api/v1")

	// Public authentication routes
	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// 2FA routes with temporary token middleware
	auth.Post("/2fa/verify", middleware.TempAuthMiddleware(config.JWTManager), authHandler.Verify2FA)

	// Protected authentication routes
	protectedAuth := auth.Group("/", middleware.AuthMiddleware(config.JWTManager))
	protectedAuth.Post("/2fa/setup", authHandler.Setup2FA)
	protectedAuth.Get("/profile", authHandler.GetProfile)

	// Public restaurant routes
	restaurants := v1.Group("/restaurants")
	restaurants.Get("/search", restaurantHandler.SearchRestaurants)
	restaurants.Get("/nearby", restaurantHandler.SearchNearbyRestaurants)
	restaurants.Get("/recommendations", restaurantHandler.GetRestaurantRecommendations)
	restaurants.Post("/format-telegram", restaurantHandler.FormatRestaurantForTelegram)

	// Telegram webhook routes
	telegram := v1.Group("/telegram")
	telegram.Post("/webhook", telegramHandler.HandleWebhook)

	// Protected admin routes (optional - for monitoring)
	admin := v1.Group("/admin", middleware.AuthMiddleware(config.JWTManager))
	admin.Get("/logs/user/:user_id", func(c *fiber.Ctx) error {
		// This would be implemented in a separate admin handler
		return c.JSON(fiber.Map{"message": "Admin endpoint - not implemented yet"})
	})

	// Swagger documentation route
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Not Found",
			"message": "The requested resource was not found",
			"path":    c.Path(),
		})
	})
}