package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"telegram-bot/internal/domain"
	"telegram-bot/internal/usecase"
	"telegram-bot/pkg/response"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authUsecase *usecase.AuthUsecase
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authUsecase *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body domain.UserRegisterRequest true "Registration request"
// @Success 201 {object} response.APIResponse{data=domain.User}
// @Failure 400 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req domain.UserRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", err.Error())
	}

	// Basic validation
	if req.Email == "" {
		return response.BadRequest(c, "Email is required")
	}
	if req.Password == "" {
		return response.BadRequest(c, "Password is required")
	}
	if req.FullName == "" {
		return response.BadRequest(c, "Full name is required")
	}
	if len(req.Password) < 8 {
		return response.BadRequest(c, "Password must be at least 8 characters")
	}

	user, err := h.authUsecase.Register(c.Context(), &req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, "User registered successfully", user)
}

// Login handles user login
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body domain.UserLoginRequest true "Login request"
// @Success 200 {object} response.APIResponse{data=domain.AuthResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req domain.UserLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", err.Error())
	}

	// Basic validation
	if req.Email == "" {
		return response.BadRequest(c, "Email is required")
	}
	if req.Password == "" {
		return response.BadRequest(c, "Password is required")
	}

	authResp, err := h.authUsecase.Login(c.Context(), &req)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	message := "Login successful"
	if authResp.Requires2FA {
		message = "2FA verification required"
	}

	return response.Success(c, message, authResp)
}

// Setup2FA handles 2FA setup
// @Summary Setup 2FA for user
// @Description Generate TOTP secret and QR code for 2FA setup
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=domain.AuthResponse}
// @Failure 401 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /auth/2fa/setup [post]
func (h *AuthHandler) Setup2FA(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userIDValue := c.Locals("user_id")
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return response.Unauthorized(c, "Invalid user context")
	}

	authResp, err := h.authUsecase.Setup2FA(c.Context(), userID)
	if err != nil {
		return response.InternalServerError(c, err.Error())
	}

	return response.Success(c, "2FA setup successful", authResp)
}

// Verify2FA handles 2FA verification
// @Summary Verify 2FA code
// @Description Verify TOTP code and complete authentication
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body domain.Verify2FARequest true "2FA verification request"
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=domain.AuthResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /auth/2fa/verify [post]
func (h *AuthHandler) Verify2FA(c *fiber.Ctx) error {
	var req domain.Verify2FARequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", err.Error())
	}

	// Basic validation
	if req.Code == "" {
		return response.BadRequest(c, "2FA code is required")
	}
	if len(req.Code) != 6 {
		return response.BadRequest(c, "2FA code must be 6 digits")
	}

	// Get temporary token from context (set by temp auth middleware)
	tempToken, ok := c.Locals("temp_token").(string)
	if !ok {
		return response.Unauthorized(c, "Invalid temporary token")
	}

	authResp, err := h.authUsecase.Verify2FA(c.Context(), tempToken, req.Code)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	return response.Success(c, "2FA verification successful", authResp)
}

// GetProfile gets current user profile
// @Summary Get user profile
// @Description Get current authenticated user profile
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=domain.User}
// @Failure 401 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	// Get user email from context (set by auth middleware)
	userEmail, ok := c.Locals("user_email").(string)
	if !ok {
		return response.Unauthorized(c, "Invalid user context")
	}

	// For now, we'll create a simple response with the user info from JWT
	// In a real application, you might want to fetch fresh user data from the database
	userID := c.Locals("user_id").(uuid.UUID)
	is2FA := c.Locals("is_2fa").(bool)

	userProfile := map[string]interface{}{
		"id":             userID,
		"email":          userEmail,
		"is_2fa_enabled": is2FA,
	}

	return response.Success(c, "Profile retrieved successfully", userProfile)
}