# Telegram Bot API - Restaurant Finder

A RESTful API for a culinary Telegram bot built with Clean Architecture principles using Go and Fiber framework.

## Features

### 🔐 Authentication & Security

- User registration and login
- JWT-based authentication
- Two-Factor Authentication (2FA) with TOTP
- Password hashing with bcrypt
- Secure middleware for request validation

### 🍽️ Restaurant Search

- Search restaurants by name/cuisine using Google Places API
- Location-based nearby restaurant search
- Restaurant recommendations
- Detailed restaurant information with ratings, photos, and opening hours

### 🤖 Telegram Bot Integration

- Interactive Telegram bot for restaurant search
- Webhook-based message processing
- Rich message formatting with inline keyboards
- Location sharing support

### 📊 Request Logging & Monitoring

- Comprehensive request logging middleware
- IP address and user agent tracking
- Response time monitoring
- Request statistics and analytics

## Architecture

This project follows **Clean Architecture** principles with clear separation of concerns:

```
├── cmd/api/                 # Application entry point
├── config/                  # Configuration management
├── internal/
│   ├── domain/             # Business entities (User, Restaurant, etc.)
│   ├── usecase/            # Business logic layer
│   ├── repository/         # Data access layer
│   │   ├── postgres/       # PostgreSQL implementations
│   │   └── external/       # External API implementations
│   └── delivery/http/      # HTTP handlers and middleware
├── pkg/                    # Shared utilities
└── database/migrations/    # Database schema
```

## Prerequisites

### Required Software

- Go 1.21 or higher
- PostgreSQL 12+
- Git

### API Keys Required

1. **Google Places API Key**

   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or select existing one
   - Enable "Places API"
   - Create credentials (API Key)
   - Restrict the key to Places API for security
2. **Telegram Bot Token**

   - Open Telegram and search for `@BotFather`
   - Send `/newbot` command
   - Follow instructions to create your bot
   - Save the bot token provided

## Installation Steps

### 1. Clone and Setup Project

```bash
# Clone the repository
git clone <your-repo-url>
cd telegram-bot

# Install dependencies
go mod download
```

### 2. Database Setup

```bash
# Install PostgreSQL (macOS)
brew install postgresql
brew services start postgresql

# Create database
createdb telegram_bot

# Run migrations
psql -d telegram_bot -f database/migrations/001_create_tables.sql
```

### 3. Configuration Setup

Copy the example configuration file:

```bash
cp config.yaml.example config.yaml
```

Edit the `config.yaml` file with your configuration:

```bash
nano config.yaml
```

Update the configuration with your values:

```yaml
app:
  name: "Telegram Bot API"
  version: "1.0.0"
  environment: "development"
  debug: true

server:
  host: "localhost"
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "your_postgres_password"
  name: "telegram_bot"
  sslmode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300s

jwt:
  secret: "your-super-secret-jwt-key-here"
  expiry: 24h
  refresh_expiry: 168h

google:
  places_api_key: "your-google-places-api-key"
  timeout: 10s

telegram:
  bot_token: "your-telegram-bot-token"
  webhook_url: "https://your-domain.com/api/v1/telegram/webhook"
  timeout: 30s

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

logging:
  level: "info"
  format: "json"
  output: "stdout"

rate_limit:
  requests_per_minute: 60
  burst: 10

cors:
  allowed_origins:
    - "http://localhost:3000"
    - "https://your-frontend-domain.com"
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allowed_headers:
    - "Content-Type"
    - "Authorization"
  allow_credentials: true
```

## API Endpoints

### Authentication

- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/2fa/setup` - Setup 2FA (protected)
- `POST /api/v1/auth/2fa/verify` - Verify 2FA code
- `GET /api/v1/auth/profile` - Get user profile (protected)

### Restaurants

- `GET /api/v1/restaurants/search` - Search restaurants
- `GET /api/v1/restaurants/nearby` - Find nearby restaurants
- `GET /api/v1/restaurants/recommendations` - Get recommendations
- `POST /api/v1/restaurants/format-telegram` - Format for Telegram

### Telegram

- `POST /api/v1/telegram/webhook` - Telegram webhook endpoint

### Health Check

- `GET /health` - Service health status

## Telegram Bot Commands

- `/start` - Welcome message and instructions
- `/help` - Show available commands
- `/search <query>` - Search for restaurants
- `/nearby` - Instructions for location-based search

## Running the Application

### Using CLI Commands

The application now includes a comprehensive CLI tool:

```bash
# Build both API server and CLI tool
make build

# Start the server using CLI
./bin/cli server

# Or using Makefile shortcut
make server

# Run with custom configuration
./bin/cli server --config custom-config.yaml

# Run with custom port and auto-migration
./bin/cli server --port 9090 --auto-migrate
```

### Database Migration

```bash
# Run database migrations
./bin/cli migrate up

# Check migration status
./bin/cli migrate status

# Using Makefile
make db-migrate
make db-status
```

### Development Mode

```bash
# Run in development mode with live reload (requires air)
make dev
# Run migrations
./bin/cli migrate up

# Check migration status
./bin/cli migrate status

# Rollback last migration
./bin/cli migrate down

# Reset database (WARNING: deletes all data)
./bin/cli migrate reset --force

# Using Makefile
make db-migrate
make db-status
make db-migrate-down
make db-reset
```

### Testing

```bash
# Run unit tests
./bin/cli test unit

# Run integration tests
./bin/cli test integration

# Run tests with coverage
./bin/cli test coverage --html

# Run benchmark tests
./bin/cli test benchmark

# Run all tests
./bin/cli test all

# Using Makefile
make test
make test-all
make test-coverage
make test-integration
make test-benchmark
```

### Development

```bash
# Development with live reload
make dev

# Build application
make build

# Run linting
make lint
```

### Building for Production

```bash
go build -o bin/api cmd/api/main.go
```

### Docker Support

```bash
# Build image
docker build -t telegram-bot-api .

# Run with docker-compose
docker-compose up -d
```

## API Documentation

Swagger documentation is available at `/swagger/` when running in development mode.
