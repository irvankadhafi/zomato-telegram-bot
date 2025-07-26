package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"telegram-bot/internal/domain"
)

// GooglePlacesRepository implements RestaurantRepository for Google Places API
type GooglePlacesRepository struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewGooglePlacesRepository creates a new Google Places repository
func NewGooglePlacesRepository(apiKey string) *GooglePlacesRepository {
	return &GooglePlacesRepository{
		apiKey:  apiKey,
		baseURL: "https://maps.googleapis.com/maps/api/place",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchByName searches restaurants by name using Google Places Text Search API
func (r *GooglePlacesRepository) SearchByName(ctx context.Context, query string, location string) (*domain.RestaurantSearchResponse, error) {
	// Build URL
	u, err := url.Parse(fmt.Sprintf("%s/textsearch/json", r.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	params := url.Values{}
	params.Add("query", query+" restaurant")
	params.Add("key", r.apiKey)
	params.Add("type", "restaurant")

	if location != "" {
		params.Add("location", location)
		params.Add("radius", "5000") // 5km radius
	}

	u.RawQuery = params.Encode()

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var googleResp domain.GooglePlacesResponse
	err = json.NewDecoder(resp.Body).Decode(&googleResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check API response status
	if googleResp.Status != "OK" && googleResp.Status != "ZERO_RESULTS" {
		return nil, fmt.Errorf("Google Places API error: %s - %s", googleResp.Status, googleResp.ErrorMessage)
	}

	// Convert to domain response
	return &domain.RestaurantSearchResponse{
		Restaurants:   googleResp.Results,
		Status:        googleResp.Status,
		NextPageToken: googleResp.NextPageToken,
	}, nil
}

// SearchNearby searches restaurants near a location using Google Places Nearby Search API
func (r *GooglePlacesRepository) SearchNearby(ctx context.Context, lat, lng float64, radius int, restaurantType string) (*domain.RestaurantSearchResponse, error) {
	// Build URL
	u, err := url.Parse(fmt.Sprintf("%s/nearbysearch/json", r.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	params := url.Values{}
	params.Add("location", fmt.Sprintf("%f,%f", lat, lng))
	params.Add("radius", strconv.Itoa(radius))
	params.Add("type", "restaurant")
	params.Add("key", r.apiKey)

	if restaurantType != "" {
		params.Add("keyword", restaurantType)
	}

	u.RawQuery = params.Encode()

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var googleResp domain.GooglePlacesResponse
	err = json.NewDecoder(resp.Body).Decode(&googleResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check API response status
	if googleResp.Status != "OK" && googleResp.Status != "ZERO_RESULTS" {
		return nil, fmt.Errorf("Google Places API error: %s - %s", googleResp.Status, googleResp.ErrorMessage)
	}

	// Convert to domain response
	return &domain.RestaurantSearchResponse{
		Restaurants:   googleResp.Results,
		Status:        googleResp.Status,
		NextPageToken: googleResp.NextPageToken,
	}, nil
}

// GetPlaceDetails gets detailed information about a place (optional enhancement)
func (r *GooglePlacesRepository) GetPlaceDetails(ctx context.Context, placeID string) (*domain.Restaurant, error) {
	// Build URL
	u, err := url.Parse(fmt.Sprintf("%s/details/json", r.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	params := url.Values{}
	params.Add("place_id", placeID)
	params.Add("key", r.apiKey)
	params.Add("fields", "name,formatted_address,rating,user_ratings_total,price_level,opening_hours,geometry,photos,types,business_status")

	u.RawQuery = params.Encode()

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var detailResp struct {
		Result domain.Restaurant `json:"result"`
		Status string           `json:"status"`
		ErrorMessage string     `json:"error_message,omitempty"`
	}

	err = json.NewDecoder(resp.Body).Decode(&detailResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check API response status
	if detailResp.Status != "OK" {
		return nil, fmt.Errorf("Google Places API error: %s - %s", detailResp.Status, detailResp.ErrorMessage)
	}

	return &detailResp.Result, nil
}