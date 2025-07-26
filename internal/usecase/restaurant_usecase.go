package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"telegram-bot/internal/domain"
)

// RestaurantRepository defines the interface for restaurant data operations
type RestaurantRepository interface {
	SearchByName(ctx context.Context, query string, location string) (*domain.RestaurantSearchResponse, error)
	SearchNearby(ctx context.Context, lat, lng float64, radius int, restaurantType string) (*domain.RestaurantSearchResponse, error)
}

// RestaurantUsecase handles restaurant business logic
type RestaurantUsecase struct {
	restaurantRepo RestaurantRepository
}

// NewRestaurantUsecase creates a new restaurant usecase
func NewRestaurantUsecase(restaurantRepo RestaurantRepository) *RestaurantUsecase {
	return &RestaurantUsecase{
		restaurantRepo: restaurantRepo,
	}
}

// SearchByName searches restaurants by name/query
func (r *RestaurantUsecase) SearchByName(ctx context.Context, req *domain.RestaurantSearchRequest) (*domain.RestaurantSearchResponse, error) {
	// Validate input
	if strings.TrimSpace(req.Query) == "" {
		return nil, errors.New("search query cannot be empty")
	}

	// Clean and validate query
	query := strings.TrimSpace(req.Query)
	if len(query) < 2 {
		return nil, errors.New("search query must be at least 2 characters")
	}

	// Perform search
	result, err := r.restaurantRepo.SearchByName(ctx, query, req.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to search restaurants: %w", err)
	}

	// Filter and enhance results
	filteredRestaurants := r.filterAndEnhanceResults(result.Restaurants)

	return &domain.RestaurantSearchResponse{
		Restaurants:   filteredRestaurants,
		Status:        result.Status,
		NextPageToken: result.NextPageToken,
	}, nil
}

// SearchNearby searches restaurants near a location
func (r *RestaurantUsecase) SearchNearby(ctx context.Context, lat, lng float64, radius int, restaurantType string) (*domain.RestaurantSearchResponse, error) {
	// Validate coordinates
	if lat < -90 || lat > 90 {
		return nil, errors.New("invalid latitude")
	}
	if lng < -180 || lng > 180 {
		return nil, errors.New("invalid longitude")
	}

	// Set default radius if not provided
	if radius <= 0 {
		radius = 1000 // 1km default
	}
	if radius > 50000 {
		radius = 50000 // 50km max
	}

	// Perform search
	result, err := r.restaurantRepo.SearchNearby(ctx, lat, lng, radius, restaurantType)
	if err != nil {
		return nil, fmt.Errorf("failed to search nearby restaurants: %w", err)
	}

	// Filter and enhance results
	filteredRestaurants := r.filterAndEnhanceResults(result.Restaurants)

	return &domain.RestaurantSearchResponse{
		Restaurants:   filteredRestaurants,
		Status:        result.Status,
		NextPageToken: result.NextPageToken,
	}, nil
}

// filterAndEnhanceResults filters and enhances restaurant results
func (r *RestaurantUsecase) filterAndEnhanceResults(restaurants []domain.Restaurant) []domain.Restaurant {
	var filtered []domain.Restaurant

	for _, restaurant := range restaurants {
		// Skip restaurants with no name
		if strings.TrimSpace(restaurant.Name) == "" {
			continue
		}

		// Skip permanently closed restaurants
		if restaurant.BusinessStatus == "CLOSED_PERMANENTLY" {
			continue
		}

		// Enhance restaurant data
		enhanced := r.enhanceRestaurantData(restaurant)
		filtered = append(filtered, enhanced)
	}

	// Limit results to top 10
	if len(filtered) > 10 {
		filtered = filtered[:10]
	}

	return filtered
}

// enhanceRestaurantData enhances restaurant data with additional processing
func (r *RestaurantUsecase) enhanceRestaurantData(restaurant domain.Restaurant) domain.Restaurant {
	// Clean up name
	restaurant.Name = strings.TrimSpace(restaurant.Name)

	// Clean up address
	restaurant.FormattedAddress = strings.TrimSpace(restaurant.FormattedAddress)

	// Ensure rating is within valid range
	if restaurant.Rating < 0 {
		restaurant.Rating = 0
	}
	if restaurant.Rating > 5 {
		restaurant.Rating = 5
	}

	// Set default business status if empty
	if restaurant.BusinessStatus == "" {
		restaurant.BusinessStatus = "OPERATIONAL"
	}

	return restaurant
}

// FormatRestaurantForTelegram formats restaurant data for Telegram message
func (r *RestaurantUsecase) FormatRestaurantForTelegram(restaurant domain.Restaurant) string {
	var message strings.Builder

	// Restaurant name
	message.WriteString(fmt.Sprintf("🍽️ **%s**\n", restaurant.Name))

	// Rating
	if restaurant.Rating > 0 {
		stars := r.generateStars(restaurant.Rating)
		message.WriteString(fmt.Sprintf("⭐ %s (%.1f/5)\n", stars, restaurant.Rating))
		if restaurant.UserRatingsTotal > 0 {
			message.WriteString(fmt.Sprintf("👥 %d reviews\n", restaurant.UserRatingsTotal))
		}
	}

	// Address
	if restaurant.FormattedAddress != "" {
		message.WriteString(fmt.Sprintf("📍 %s\n", restaurant.FormattedAddress))
	}

	// Price level
	if restaurant.PriceLevel > 0 {
		priceSymbols := strings.Repeat("💰", restaurant.PriceLevel)
		message.WriteString(fmt.Sprintf("💵 %s\n", priceSymbols))
	}

	// Opening hours
	if restaurant.OpeningHours != nil {
		if restaurant.OpeningHours.OpenNow {
			message.WriteString("🟢 Open now\n")
		} else {
			message.WriteString("🔴 Closed now\n")
		}
	}

	// Business status
	switch restaurant.BusinessStatus {
	case "OPERATIONAL":
		message.WriteString("✅ Operational\n")
	case "CLOSED_TEMPORARILY":
		message.WriteString("⏸️ Temporarily closed\n")
	case "CLOSED_PERMANENTLY":
		message.WriteString("❌ Permanently closed\n")
	}

	return message.String()
}

// generateStars generates star representation for rating
func (r *RestaurantUsecase) generateStars(rating float64) string {
	fullStars := int(rating)
	hasHalfStar := rating-float64(fullStars) >= 0.5

	var stars strings.Builder
	for i := 0; i < fullStars; i++ {
		stars.WriteString("⭐")
	}
	if hasHalfStar {
		stars.WriteString("⭐")
	}

	return stars.String()
}