package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"

	"telegram-bot/internal/domain"
	"telegram-bot/pkg/crypto"
	"telegram-bot/pkg/jwt"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	UpdateTOTPSecret(ctx context.Context, userID uuid.UUID, secret string) error
	Enable2FA(ctx context.Context, userID uuid.UUID) error
}

// AuthUsecase handles authentication business logic
type AuthUsecase struct {
	userRepo   UserRepository
	jwtManager *jwt.JWTManager
}

// NewAuthUsecase creates a new auth usecase
func NewAuthUsecase(userRepo UserRepository, jwtManager *jwt.JWTManager) *AuthUsecase {
	return &AuthUsecase{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register registers a new user
func (a *AuthUsecase) Register(ctx context.Context, req *domain.UserRegisterRequest) (*domain.User, error) {
	// Check if user already exists
	existingUser, err := a.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &domain.User{
		ID:           uuid.New(),
		Email:        req.Email,
		Password:     hashedPassword,
		FullName:     req.FullName,
		IsActive:     true,
		Is2FAEnabled: false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = a.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Don't return password
	user.Password = ""
	return user, nil
}

// Login authenticates a user
func (a *AuthUsecase) Login(ctx context.Context, req *domain.UserLoginRequest) (*domain.AuthResponse, error) {
	// Get user by email
	user, err := a.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check password
	if !crypto.CheckPassword(req.Password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is disabled")
	}

	// If 2FA is enabled, return temporary token
	if user.Is2FAEnabled {
		tempToken, err := a.jwtManager.GenerateTempToken(user.ID, user.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to generate temp token: %w", err)
		}

		return &domain.AuthResponse{
			TempToken:   tempToken,
			Requires2FA: true,
		}, nil
	}

	// Generate JWT token
	token, err := a.jwtManager.GenerateToken(user.ID, user.Email, user.Is2FAEnabled)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Don't return password
	user.Password = ""

	return &domain.AuthResponse{
		Token:       token,
		Requires2FA: false,
		User:        user,
	}, nil
}

// Setup2FA sets up 2FA for a user
func (a *AuthUsecase) Setup2FA(ctx context.Context, userID uuid.UUID) (*domain.AuthResponse, error) {
	// Get user
	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Generate TOTP secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Telegram Bot API",
		AccountName: user.Email,
		SecretSize:  32,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	// Save TOTP secret
	err = a.userRepo.UpdateTOTPSecret(ctx, userID, key.Secret())
	if err != nil {
		return nil, fmt.Errorf("failed to save TOTP secret: %w", err)
	}

	// Generate QR code URL
	qrCodeURL := key.URL()

	return &domain.AuthResponse{
		QRCodeURL:   qrCodeURL,
		BackupCodes: key.Secret(), // In production, generate proper backup codes
	}, nil
}

// Verify2FA verifies 2FA code and completes authentication
func (a *AuthUsecase) Verify2FA(ctx context.Context, tempToken string, code string) (*domain.AuthResponse, error) {
	// Validate temporary token
	claims, err := a.jwtManager.ValidateTempToken(tempToken)
	if err != nil {
		return nil, errors.New("invalid or expired temporary token")
	}

	// Get user
	user, err := a.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Verify TOTP code
	valid := totp.Validate(code, user.TOTPSecret)
	if !valid {
		return nil, errors.New("invalid 2FA code")
	}

	// Enable 2FA if not already enabled
	if !user.Is2FAEnabled {
		err = a.userRepo.Enable2FA(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to enable 2FA: %w", err)
		}
		user.Is2FAEnabled = true
	}

	// Generate final JWT token
	token, err := a.jwtManager.GenerateToken(user.ID, user.Email, user.Is2FAEnabled)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Don't return password and TOTP secret
	user.Password = ""
	user.TOTPSecret = ""

	return &domain.AuthResponse{
		Token:       token,
		Requires2FA: false,
		User:        user,
	}, nil
}
