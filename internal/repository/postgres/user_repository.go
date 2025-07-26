package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"telegram-bot/internal/domain"
)

// UserRepository implements the UserRepository interface for PostgreSQL
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password, full_name, is_active, is_2fa_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Password,
		user.FullName,
		user.IsActive,
		user.Is2FAEnabled,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByEmail gets a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password, full_name, is_active, is_2fa_enabled, totp_secret, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	var totpSecret *string

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.FullName,
		&user.IsActive,
		&user.Is2FAEnabled,
		&totpSecret,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if totpSecret != nil {
		user.TOTPSecret = *totpSecret
	}

	return &user, nil
}

// GetByID gets a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, password, full_name, is_active, is_2fa_enabled, totp_secret, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	var totpSecret *string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.FullName,
		&user.IsActive,
		&user.Is2FAEnabled,
		&totpSecret,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	if totpSecret != nil {
		user.TOTPSecret = *totpSecret
	}

	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $2, password = $3, full_name = $4, is_active = $5, is_2fa_enabled = $6, updated_at = $7
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Password,
		user.FullName,
		user.IsActive,
		user.Is2FAEnabled,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdateTOTPSecret updates the TOTP secret for a user
func (r *UserRepository) UpdateTOTPSecret(ctx context.Context, userID uuid.UUID, secret string) error {
	query := `
		UPDATE users
		SET totp_secret = $2, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID, secret)
	if err != nil {
		return fmt.Errorf("failed to update TOTP secret: %w", err)
	}

	return nil
}

// Enable2FA enables 2FA for a user
func (r *UserRepository) Enable2FA(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET is_2fa_enabled = true, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	return nil
}