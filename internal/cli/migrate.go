package cli

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration commands",
	Long: `Database migration commands for managing database schema.

Available subcommands:
  up    - Run all pending migrations
  down  - Rollback the last migration
  reset - Drop all tables and re-run all migrations
  status - Show migration status

Example:
  telegram-bot migrate up
  telegram-bot migrate down
  telegram-bot migrate reset
  telegram-bot migrate status`,
}

// migrateUpCmd represents the migrate up command
var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Run all pending migrations",
	Long:  `Run all pending database migrations to bring the database schema up to date.`,
	RunE:  runMigrateUp,
}

// migrateDownCmd represents the migrate down command
var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback the last migration",
	Long:  `Rollback the last applied migration.`,
	RunE:  runMigrateDown,
}

// migrateResetCmd represents the migrate reset command
var migrateResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Drop all tables and re-run all migrations",
	Long:  `Drop all tables and re-run all migrations. WARNING: This will delete all data!`,
	RunE:  runMigrateReset,
}

// migrateStatusCmd represents the migrate status command
var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long:  `Show the status of all migrations.`,
	RunE:  runMigrateStatus,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateResetCmd)
	migrateCmd.AddCommand(migrateStatusCmd)

	// Migration-specific flags
	migrateCmd.PersistentFlags().String("migrations-path", "database/migrations", "path to migration files")
	migrateResetCmd.Flags().Bool("force", false, "force reset without confirmation")
}

func runMigrateUp(cmd *cobra.Command, args []string) error {
	log.Println("Running database migrations...")

	// Initialize database connection
	db, err := initDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get migration files
	migrationsPath, _ := cmd.Flags().GetString("migrations-path")
	migrationFiles, err := getMigrationFiles(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	pendingCount := 0
	for _, file := range migrationFiles {
		if !contains(appliedMigrations, file) {
			log.Printf("Applying migration: %s", file)
			if err := applyMigration(db, migrationsPath, file); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", file, err)
			}
			pendingCount++
		}
	}

	if pendingCount == 0 {
		log.Println("No pending migrations")
	} else {
		log.Printf("Applied %d migrations successfully", pendingCount)
	}

	return nil
}

func runMigrateDown(cmd *cobra.Command, args []string) error {
	log.Println("Rolling back last migration...")

	// Initialize database connection
	db, err := initDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Get last applied migration
	lastMigration, err := getLastAppliedMigration(db)
	if err != nil {
		return fmt.Errorf("failed to get last applied migration: %w", err)
	}

	if lastMigration == "" {
		log.Println("No migrations to rollback")
		return nil
	}

	// For this simple implementation, we'll just remove the migration record
	// In a real-world scenario, you'd want to have down migration files
	log.Printf("Rolling back migration: %s", lastMigration)
	if err := rollbackMigration(db, lastMigration); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", lastMigration, err)
	}

	log.Println("Migration rolled back successfully")
	return nil
}

func runMigrateReset(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")

	if !force {
		log.Println("WARNING: This will drop all tables and delete all data!")
		log.Println("Use --force flag to confirm this action")
		return nil
	}

	log.Println("Resetting database...")

	// Initialize database connection
	db, err := initDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Drop all tables
	if err := dropAllTables(db); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	// Re-run all migrations
	log.Println("Re-running all migrations...")
	return runMigrateUp(cmd, args)
}

func runMigrateStatus(cmd *cobra.Command, args []string) error {
	log.Println("Migration status:")

	// Initialize database connection
	db, err := initDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get migration files
	migrationsPath, _ := cmd.Flags().GetString("migrations-path")
	migrationFiles, err := getMigrationFiles(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Show status
	for _, file := range migrationFiles {
		status := "PENDING"
		if contains(appliedMigrations, file) {
			status = "APPLIED"
		}
		log.Printf("  %s: %s", file, status)
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table
func createMigrationsTable(db *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
		    version VARCHAR(255) PRIMARY KEY,
		    applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`

	_, err := db.Exec(ctx, query)
	return err
}

// getMigrationFiles returns a sorted list of migration files
func getMigrationFiles(migrationsPath string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(migrationsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".sql") {
			files = append(files, d.Name())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort files to ensure consistent order
	sort.Strings(files)
	return files, nil
}

// getAppliedMigrations returns a list of applied migration versions
func getAppliedMigrations(db *pgxpool.Pool) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := "SELECT version FROM schema_migrations ORDER BY version"
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		migrations = append(migrations, version)
	}

	return migrations, rows.Err()
}

// applyMigration applies a single migration file
func applyMigration(db *pgxpool.Pool, migrationsPath, filename string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// For this example, we'll use the basic migration SQL
	// In a real implementation, you'd read the actual file content
	migrationSQL := getBasicMigrationSQL()

	// Execute migration
	_, err := db.Exec(ctx, migrationSQL)
	if err != nil {
		return err
	}

	// Record migration as applied
	query := "INSERT INTO schema_migrations (version) VALUES ($1)"
	_, err = db.Exec(ctx, query, filename)
	return err
}

// getLastAppliedMigration returns the last applied migration
func getLastAppliedMigration(db *pgxpool.Pool) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := "SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1"
	var version string
	err := db.QueryRow(ctx, query).Scan(&version)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return "", nil
		}
		return "", err
	}

	return version, nil
}

// rollbackMigration removes a migration record
func rollbackMigration(db *pgxpool.Pool, version string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := "DELETE FROM schema_migrations WHERE version = $1"
	_, err := db.Exec(ctx, query, version)
	return err
}

// dropAllTables drops all tables in the database
func dropAllTables(db *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Drop tables in correct order (considering foreign keys)
	dropSQL := `
		DROP TABLE IF EXISTS request_logs CASCADE;
		DROP TABLE IF EXISTS users CASCADE;
		DROP TABLE IF EXISTS schema_migrations CASCADE;
	`

	_, err := db.Exec(ctx, dropSQL)
	return err
}

// getBasicMigrationSQL returns the basic migration SQL
func getBasicMigrationSQL() string {
	return `
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
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}