package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents the user entity
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Password     string    `json:"-"` // Never expose password in JSON
	FullName     string    `json:"full_name"`
	IsActive     bool      `json:"is_active"`
	Is2FAEnabled bool      `json:"is_2fa_enabled"`
	TOTPSecret   string    `json:"-"` // Never expose TOTP secret
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserRegisterRequest represents user registration request
type UserRegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required,min=2"`
}

// UserLoginRequest represents user login request
type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Setup2FARequest represents 2FA setup request
type Setup2FARequest struct {
	UserID uuid.UUID `json:"user_id"`
}

// Verify2FARequest represents 2FA verification request
type Verify2FARequest struct {
	UserID uuid.UUID `json:"user_id"`
	Code   string    `json:"code" validate:"required,len=6"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token        string `json:"token,omitempty"`
	TempToken    string `json:"temp_token,omitempty"`
	Requires2FA  bool   `json:"requires_2fa"`
	User         *User  `json:"user,omitempty"`
	QRCodeURL    string `json:"qr_code_url,omitempty"`
	BackupCodes  string `json:"backup_codes,omitempty"`
}