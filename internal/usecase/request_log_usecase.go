package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"telegram-bot/internal/domain"
)

// RequestLogRepository defines the interface for request log data operations
type RequestLogRepository interface {
	Create(ctx context.Context, log *domain.RequestLog) error
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.RequestLog, error)
	GetByIPAddress(ctx context.Context, ipAddress string, limit, offset int) ([]domain.RequestLog, error)
	GetStats(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error)
}

// RequestLogUsecase handles request logging business logic
type RequestLogUsecase struct {
	requestLogRepo RequestLogRepository
}

// NewRequestLogUsecase creates a new request log usecase
func NewRequestLogUsecase(requestLogRepo RequestLogRepository) *RequestLogUsecase {
	return &RequestLogUsecase{
		requestLogRepo: requestLogRepo,
	}
}

// LogRequest logs a new request
func (r *RequestLogUsecase) LogRequest(ctx context.Context, req *domain.CreateRequestLogRequest) error {
	// Create request log entity
	requestLog := &domain.RequestLog{
		ID:           uuid.New(),
		IPAddress:    req.IPAddress,
		HTTPMethod:   req.HTTPMethod,
		Path:         req.Path,
		UserAgent:    req.UserAgent,
		StatusCode:   req.StatusCode,
		ResponseTime: req.ResponseTime,
		UserID:       req.UserID,
		CreatedAt:    time.Now(),
	}

	// Save to repository
	err := r.requestLogRepo.Create(ctx, requestLog)
	if err != nil {
		return fmt.Errorf("failed to log request: %w", err)
	}

	return nil
}

// GetUserRequestLogs gets request logs for a specific user
func (r *RequestLogUsecase) GetUserRequestLogs(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.RequestLog, error) {
	// Validate pagination parameters
	if limit <= 0 {
		limit = 50 // default limit
	}
	if limit > 100 {
		limit = 100 // max limit
	}
	if offset < 0 {
		offset = 0
	}

	logs, err := r.requestLogRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user request logs: %w", err)
	}

	return logs, nil
}

// GetIPRequestLogs gets request logs for a specific IP address
func (r *RequestLogUsecase) GetIPRequestLogs(ctx context.Context, ipAddress string, limit, offset int) ([]domain.RequestLog, error) {
	// Validate pagination parameters
	if limit <= 0 {
		limit = 50 // default limit
	}
	if limit > 100 {
		limit = 100 // max limit
	}
	if offset < 0 {
		offset = 0
	}

	logs, err := r.requestLogRepo.GetByIPAddress(ctx, ipAddress, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get IP request logs: %w", err)
	}

	return logs, nil
}

// GetRequestStats gets request statistics for a date range
func (r *RequestLogUsecase) GetRequestStats(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	// Validate date range
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("end date cannot be before start date")
	}

	// Limit date range to prevent excessive queries
	maxDuration := 30 * 24 * time.Hour // 30 days
	if endDate.Sub(startDate) > maxDuration {
		return nil, fmt.Errorf("date range cannot exceed 30 days")
	}

	stats, err := r.requestLogRepo.GetStats(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get request stats: %w", err)
	}

	return stats, nil
}