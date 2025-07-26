package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"telegram-bot/internal/domain"
	"telegram-bot/internal/usecase"
	"telegram-bot/pkg/response"
)

// RestaurantHandler handles restaurant endpoints
type RestaurantHandler struct {
	restaurantUsecase *usecase.RestaurantUsecase
}

// NewRestaurantHandler creates a new restaurant handler
func NewRestaurantHandler(restaurantUsecase *usecase.RestaurantUsecase) *RestaurantHandler {
	return &RestaurantHandler{
		restaurantUsecase: restaurantUsecase,
	}
}

// SearchRestaurants handles restaurant search by name/query
// @Summary Search restaurants
// @Description Search restaurants by name or query using Google Places API
// @Tags Restaurants
// @Accept json
// @Produce json
// @Param query query string true "Search query (restaurant name, cuisine, etc.)"
// @Param location query string false "Location (optional, e.g., 'New York, NY')"
// @Success 200 {object} response.APIResponse{data=domain.RestaurantSearchResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /restaurants/search [get]
func (h *RestaurantHandler) SearchRestaurants(c *fiber.Ctx) error {
	// Get query parameters
	query := c.Query("query")
	location := c.Query("location")

	// Validate query
	if query == "" {
		return response.BadRequest(c, "Query parameter is required")
	}

	// Create search request
	req := &domain.RestaurantSearchRequest{
		Query:    query,
		Location: location,
	}

	// Perform search
	result, err := h.restaurantUsecase.SearchByName(c.Context(), req)
	if err != nil {
		return response.InternalServerError(c, "Failed to search restaurants", err.Error())
	}

	return response.Success(c, "Restaurants retrieved successfully", result)
}

// SearchNearbyRestaurants handles nearby restaurant search
// @Summary Search nearby restaurants
// @Description Search restaurants near a specific location using coordinates
// @Tags Restaurants
// @Accept json
// @Produce json
// @Param lat query number true "Latitude"
// @Param lng query number true "Longitude"
// @Param radius query integer false "Search radius in meters (default: 1000, max: 50000)"
// @Param type query string false "Restaurant type/cuisine (optional)"
// @Success 200 {object} response.APIResponse{data=domain.RestaurantSearchResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /restaurants/nearby [get]
func (h *RestaurantHandler) SearchNearbyRestaurants(c *fiber.Ctx) error {
	// Get and validate coordinates
	latStr := c.Query("lat")
	lngStr := c.Query("lng")

	if latStr == "" || lngStr == "" {
		return response.BadRequest(c, "Latitude and longitude are required")
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid latitude format")
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid longitude format")
	}

	// Get optional parameters
	radius := 1000 // default 1km
	if radiusStr := c.Query("radius"); radiusStr != "" {
		if r, err := strconv.Atoi(radiusStr); err == nil {
			radius = r
		}
	}

	restaurantType := c.Query("type")

	// Perform search
	result, err := h.restaurantUsecase.SearchNearby(c.Context(), lat, lng, radius, restaurantType)
	if err != nil {
		return response.InternalServerError(c, "Failed to search nearby restaurants", err.Error())
	}

	return response.Success(c, "Nearby restaurants retrieved successfully", result)
}

// GetRestaurantRecommendations provides restaurant recommendations
// @Summary Get restaurant recommendations
// @Description Get curated restaurant recommendations based on popular searches
// @Tags Restaurants
// @Produce json
// @Param cuisine query string false "Preferred cuisine type"
// @Param location query string false "Location for recommendations"
// @Success 200 {object} response.APIResponse{data=domain.RestaurantSearchResponse}
// @Failure 500 {object} response.APIResponse
// @Router /restaurants/recommendations [get]
func (h *RestaurantHandler) GetRestaurantRecommendations(c *fiber.Ctx) error {
	cuisine := c.Query("cuisine", "restaurant")
	location := c.Query("location", "")

	// Create search request for popular/recommended restaurants
	req := &domain.RestaurantSearchRequest{
		Query:    "popular " + cuisine,
		Location: location,
	}

	// Perform search
	result, err := h.restaurantUsecase.SearchByName(c.Context(), req)
	if err != nil {
		return response.InternalServerError(c, "Failed to get restaurant recommendations", err.Error())
	}

	return response.Success(c, "Restaurant recommendations retrieved successfully", result)
}

// FormatRestaurantForTelegram formats restaurant data for Telegram
// @Summary Format restaurant for Telegram
// @Description Format restaurant information for Telegram bot message
// @Tags Restaurants
// @Accept json
// @Produce json
// @Param request body domain.Restaurant true "Restaurant data"
// @Success 200 {object} response.APIResponse{data=string}
// @Failure 400 {object} response.APIResponse
// @Router /restaurants/format-telegram [post]
func (h *RestaurantHandler) FormatRestaurantForTelegram(c *fiber.Ctx) error {
	var restaurant domain.Restaurant
	if err := c.BodyParser(&restaurant); err != nil {
		return response.BadRequest(c, "Invalid request body", err.Error())
	}

	// Format restaurant for Telegram
	formattedMessage := h.restaurantUsecase.FormatRestaurantForTelegram(restaurant)

	return response.Success(c, "Restaurant formatted successfully", map[string]string{
		"formatted_message": formattedMessage,
	})
}