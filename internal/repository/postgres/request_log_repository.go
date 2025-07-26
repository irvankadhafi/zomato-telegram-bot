package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"telegram-bot/internal/domain"
)

// RequestLogRepository implements the RequestLogRepository interface for PostgreSQL
type RequestLogRepository struct {
	db *pgxpool.Pool
}

// NewRequestLogRepository creates a new request log repository
func NewRequestLogRepository(db *pgxpool.Pool) *RequestLogRepository {
	return &RequestLogRepository{db: db}
}

// Create creates a new request log
func (r *RequestLogRepository) Create(ctx context.Context, log *domain.RequestLog) error {
	query := `
		INSERT INTO request_logs (id, ip_address, http_method, path, user_agent, status_code, response_time, user_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(ctx, query,
		log.ID,
		log.IPAddress,
		log.HTTPMethod,
		log.Path,
		log.UserAgent,
		log.StatusCode,
		log.ResponseTime,
		log.UserID,
		log.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create request log: %w", err)
	}

	return nil
}

// GetByUserID gets request logs by user ID with pagination
func (r *RequestLogRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.RequestLog, error) {
	query := `
		SELECT id, ip_address, http_method, path, user_agent, status_code, response_time, user_id, created_at
		FROM request_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get request logs by user ID: %w", err)
	}
	defer rows.Close()

	var logs []domain.RequestLog
	for rows.Next() {
		var log domain.RequestLog
		err := rows.Scan(
			&log.ID,
			&log.IPAddress,
			&log.HTTPMethod,
			&log.Path,
			&log.UserAgent,
			&log.StatusCode,
			&log.ResponseTime,
			&log.UserID,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan request log: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating request logs: %w", err)
	}

	return logs, nil
}

// GetByIPAddress gets request logs by IP address with pagination
func (r *RequestLogRepository) GetByIPAddress(ctx context.Context, ipAddress string, limit, offset int) ([]domain.RequestLog, error) {
	query := `
		SELECT id, ip_address, http_method, path, user_agent, status_code, response_time, user_id, created_at
		FROM request_logs
		WHERE ip_address = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, ipAddress, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get request logs by IP address: %w", err)
	}
	defer rows.Close()

	var logs []domain.RequestLog
	for rows.Next() {
		var log domain.RequestLog
		err := rows.Scan(
			&log.ID,
			&log.IPAddress,
			&log.HTTPMethod,
			&log.Path,
			&log.UserAgent,
			&log.StatusCode,
			&log.ResponseTime,
			&log.UserID,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan request log: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating request logs: %w", err)
	}

	return logs, nil
}

// GetStats gets request statistics for a date range
func (r *RequestLogRepository) GetStats(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total requests
	totalQuery := `
		SELECT COUNT(*) as total_requests
		FROM request_logs
		WHERE created_at >= $1 AND created_at <= $2
	`
	var totalRequests int64
	err := r.db.QueryRow(ctx, totalQuery, startDate, endDate).Scan(&totalRequests)
	if err != nil {
		return nil, fmt.Errorf("failed to get total requests: %w", err)
	}
	stats["total_requests"] = totalRequests

	// Requests by status code
	statusQuery := `
		SELECT status_code, COUNT(*) as count
		FROM request_logs
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY status_code
		ORDER BY count DESC
	`
	statusRows, err := r.db.Query(ctx, statusQuery, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get status code stats: %w", err)
	}
	defer statusRows.Close()

	statusCodes := make(map[string]int64)
	for statusRows.Next() {
		var statusCode int
		var count int64
		err := statusRows.Scan(&statusCode, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status code stats: %w", err)
		}
		statusCodes[fmt.Sprintf("%d", statusCode)] = count
	}
	stats["status_codes"] = statusCodes

	// Requests by method
	methodQuery := `
		SELECT http_method, COUNT(*) as count
		FROM request_logs
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY http_method
		ORDER BY count DESC
	`
	methodRows, err := r.db.Query(ctx, methodQuery, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get method stats: %w", err)
	}
	defer methodRows.Close()

	methods := make(map[string]int64)
	for methodRows.Next() {
		var method string
		var count int64
		err := methodRows.Scan(&method, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan method stats: %w", err)
		}
		methods[method] = count
	}
	stats["methods"] = methods

	// Average response time
	avgTimeQuery := `
		SELECT AVG(response_time) as avg_response_time
		FROM request_logs
		WHERE created_at >= $1 AND created_at <= $2
	`
	var avgResponseTime *float64
	err = r.db.QueryRow(ctx, avgTimeQuery, startDate, endDate).Scan(&avgResponseTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get average response time: %w", err)
	}
	if avgResponseTime != nil {
		stats["avg_response_time_ms"] = *avgResponseTime
	} else {
		stats["avg_response_time_ms"] = 0
	}

	// Top paths
	pathQuery := `
		SELECT path, COUNT(*) as count
		FROM request_logs
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY path
		ORDER BY count DESC
		LIMIT 10
	`
	pathRows, err := r.db.Query(ctx, pathQuery, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get path stats: %w", err)
	}
	defer pathRows.Close()

	topPaths := make([]map[string]interface{}, 0)
	for pathRows.Next() {
		var path string
		var count int64
		err := pathRows.Scan(&path, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan path stats: %w", err)
		}
		topPaths = append(topPaths, map[string]interface{}{
			"path":  path,
			"count": count,
		})
	}
	stats["top_paths"] = topPaths

	return stats, nil
}