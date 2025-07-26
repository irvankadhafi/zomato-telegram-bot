package cli

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
	"github.com/spf13/cobra"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"telegram-bot/config"
	"telegram-bot/internal/delivery/http/router"
	"telegram-bot/internal/repository/external"
	"telegram-bot/internal/repository/postgres"
	"telegram-bot/internal/usecase"
	"telegram-bot/pkg/jwt"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP server",
	Long: `Start the HTTP server with all endpoints.

This command will:
- Initialize database connection
- Setup Telegram bot
- Start HTTP server with all routes
- Handle graceful shutdown

Example:
  telegram-bot server
  telegram-bot server --port 8080
  telegram-bot server --config custom-config.yaml`,
	RunE: runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Server-specific flags
	serverCmd.Flags().Int("port", 8080, "server port")
	serverCmd.Flags().String("host", "localhost", "server host")
	serverCmd.Flags().Bool("auto-migrate", false, "run database migrations on startup")
}

func runServer(cmd *cobra.Command, args []string) error {
	// Get flags
	port, _ := cmd.Flags().GetInt("port")
	host, _ := cmd.Flags().GetString("host")
	autoMigrate, _ := cmd.Flags().GetBool("auto-migrate")

	// Override config with flags if provided
	if cmd.Flags().Changed("port") {
		cfg.Server.Port = port
	}
	if cmd.Flags().Changed("host") {
		cfg.Server.Host = host
	}

	log.Printf("Starting Telegram Bot API Server...")
	log.Printf("Environment: %s", cfg.App.Environment)
	log.Printf("Version: %s", cfg.App.Version)

	// Initialize database connection
	db, err := initDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Run migrations if requested
	if autoMigrate {
		log.Println("Running database migrations...")
		if err := runMigrations(db); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
		log.Println("Database migrations completed")
	}

	// Initialize Telegram bot
	bot, err := initTelegramBot(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize Telegram bot: %w", err)
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
		AppName:      cfg.App.Name,
		ServerHeader: "Fiber",
		ErrorHandler: customErrorHandler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
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
	log.Printf("Server starting on %s", serverAddr)
	log.Printf("Telegram Bot: @%s", bot.Self.UserName)
	log.Printf("API Documentation: http://%s/swagger/", serverAddr)

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
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server exited gracefully")
	return nil
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
	bot.Debug = cfg.App.Debug

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

// runMigrations runs database migrations
func runMigrations(db *pgxpool.Pool) error {
	// This is a simple migration runner
	// In production, you might want to use a more sophisticated migration tool
	migrationSQL := `
		-- Create users table
		CREATE TABLE IF NOT EXISTS users (
		    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		    email VARCHAR(255) UNIQUE NOT NULL,
		    password VARCHAR(255) NOT NULL,
		    full_name VARCHAR(255) NOT NULL,
		    is_active BOOLEAN DEFAULT true,
		    is_2fa_enabled BOOLEAN DEFAULT false,
		    totp_secret VARCHAR(255),
		    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		-- Create request_logs table
		CREATE TABLE IF NOT EXISTS request_logs (
		    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		    ip_address INET NOT NULL,
		    http_method VARCHAR(10) NOT NULL,
		    path TEXT NOT NULL,
		    user_agent TEXT,
		    status_code INTEGER NOT NULL,
		    response_time BIGINT NOT NULL,
		    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
		    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		-- Create indexes
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_request_logs_user_id ON request_logs(user_id);
		CREATE INDEX IF NOT EXISTS idx_request_logs_created_at ON request_logs(created_at);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := db.Exec(ctx, migrationSQL)
	return err
}