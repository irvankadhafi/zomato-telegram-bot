package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"telegram-bot/config"
	"telegram-bot/internal/delivery/http/router"
	"telegram-bot/internal/repository/external"
	"telegram-bot/internal/repository/postgres"
	"telegram-bot/internal/usecase"
	"telegram-bot/pkg/jwt"
)

// @title Telegram Bot API
// @version 1.0
// @description A RESTful API for a culinary Telegram bot with Clean Architecture
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize Telegram bot
	bot, err := initTelegramBot(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Telegram bot: %v", err)
	}

	// Initialize JWT manager
	jwtManager := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiry)

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	requestLogRepo := postgres.NewRequestLogRepository(db)
	googlePlacesRepo := external.NewGooglePlacesRepository(cfg.Google.PlacesAPIKey)

	// Initialize use cases
	authUsecase := usecase.NewAuthUsecase(userRepo, jwtManager)
	restaurantUsecase := usecase.NewRestaurantUsecase(googlePlacesRepo)
	requestLogUsecase := usecase.NewRequestLogUsecase(requestLogRepo)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Telegram Bot API",
		ServerHeader: "Fiber",
		ErrorHandler: customErrorHandler,
	})

	// Add recovery middleware
	app.Use(recover.New())

	// Setup routes
	routerConfig := &router.RouterConfig{
		App:               app,
		AuthUsecase:       authUsecase,
		RestaurantUsecase: restaurantUsecase,
		RequestLogUsecase: requestLogUsecase,
		JWTManager:        jwtManager,
		TelegramBot:       bot,
	}
	router.SetupRoutes(routerConfig)

	// Start server
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", serverAddr)
	log.Printf("Environment: %s", cfg.App.Environment)
	log.Printf("Telegram Bot: @%s", bot.Self.UserName)

	// Graceful shutdown
	go func() {
		if err := app.Listen(serverAddr); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Gracefully shutdown the server with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// initDatabase initializes the database connection pool
func initDatabase(cfg *config.Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.Database.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established")
	return db, nil
}

// initTelegramBot initializes the Telegram bot
func initTelegramBot(cfg *config.Config) (*tgbotapi.BotAPI, error) {
	if cfg.Telegram.BotToken == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
	}

	// Set debug mode based on environment
	bot.Debug = cfg.App.Environment == "development"

	log.Printf("Authorized on account %s", bot.Self.UserName)
	return bot, nil
}

// customErrorHandler handles Fiber errors
func customErrorHandler(c *fiber.Ctx, err error) error {
	// Status code defaults to 500
	code := fiber.StatusInternalServerError

	// Retrieve the custom status code if it's a *fiber.Error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Send custom error page
	return c.Status(code).JSON(fiber.Map{
		"error":   true,
		"message": err.Error(),
		"code":    code,
	})
}