package response

import (
	"github.com/gofiber/fiber/v2"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool            `json:"success"`
	Message    string          `json:"message"`
	Data       interface{}     `json:"data"`
	Pagination PaginationMeta  `json:"pagination"`
	Error      *ErrorDetail    `json:"error,omitempty"`
}

// Success sends a successful response
func Success(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created sends a created response
func Created(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// BadRequest sends a bad request response
func BadRequest(c *fiber.Ctx, message string, details ...string) error {
	errorDetail := &ErrorDetail{
		Code:    "BAD_REQUEST",
		Message: message,
	}
	if len(details) > 0 {
		errorDetail.Details = details[0]
	}
	return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
		Success: false,
		Message: "Bad Request",
		Error:   errorDetail,
	})
}

// Unauthorized sends an unauthorized response
func Unauthorized(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
		Success: false,
		Message: "Unauthorized",
		Error: &ErrorDetail{
			Code:    "UNAUTHORIZED",
			Message: message,
		},
	})
}

// Forbidden sends a forbidden response
func Forbidden(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusForbidden).JSON(APIResponse{
		Success: false,
		Message: "Forbidden",
		Error: &ErrorDetail{
			Code:    "FORBIDDEN",
			Message: message,
		},
	})
}

// NotFound sends a not found response
func NotFound(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusNotFound).JSON(APIResponse{
		Success: false,
		Message: "Not Found",
		Error: &ErrorDetail{
			Code:    "NOT_FOUND",
			Message: message,
		},
	})
}

// InternalServerError sends an internal server error response
func InternalServerError(c *fiber.Ctx, message string, details ...string) error {
	errorDetail := &ErrorDetail{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: message,
	}
	if len(details) > 0 {
		errorDetail.Details = details[0]
	}
	return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
		Success: false,
		Message: "Internal Server Error",
		Error:   errorDetail,
	})
}

// Paginated sends a paginated response
func Paginated(c *fiber.Ctx, message string, data interface{}, pagination PaginationMeta) error {
	return c.Status(fiber.StatusOK).JSON(PaginatedResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	})
}