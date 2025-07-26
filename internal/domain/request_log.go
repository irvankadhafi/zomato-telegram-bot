package domain

import (
	"time"

	"github.com/google/uuid"
)

// RequestLog represents the request log entity
type RequestLog struct {
	ID           uuid.UUID `json:"id"`
	IPAddress    string    `json:"ip_address"`
	HTTPMethod   string    `json:"http_method"`
	Path         string    `json:"path"`
	UserAgent    string    `json:"user_agent"`
	StatusCode   int       `json:"status_code"`
	ResponseTime int64     `json:"response_time"` // in milliseconds
	UserID       *uuid.UUID `json:"user_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateRequestLogRequest represents request log creation request
type CreateRequestLogRequest struct {
	IPAddress    string     `json:"ip_address"`
	HTTPMethod   string     `json:"http_method"`
	Path         string     `json:"path"`
	UserAgent    string     `json:"user_agent"`
	StatusCode   int        `json:"status_code"`
	ResponseTime int64      `json:"response_time"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
}