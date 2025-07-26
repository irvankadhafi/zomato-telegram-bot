package domain

// Restaurant represents the restaurant entity from Google Places API
type Restaurant struct {
	PlaceID          string    `json:"place_id"`
	Name             string    `json:"name"`
	FormattedAddress string    `json:"formatted_address"`
	Rating           float64   `json:"rating"`
	UserRatingsTotal int       `json:"user_ratings_total"`
	PriceLevel       int       `json:"price_level"`
	OpeningHours     *OpeningHours `json:"opening_hours,omitempty"`
	Geometry         *Geometry `json:"geometry,omitempty"`
	Photos           []Photo   `json:"photos,omitempty"`
	Types            []string  `json:"types"`
	BusinessStatus   string    `json:"business_status"`
}

// OpeningHours represents restaurant opening hours
type OpeningHours struct {
	OpenNow     bool     `json:"open_now"`
	WeekdayText []string `json:"weekday_text"`
}

// Geometry represents location coordinates
type Geometry struct {
	Location *Location `json:"location"`
}

// Location represents latitude and longitude
type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Photo represents restaurant photo
type Photo struct {
	PhotoReference string   `json:"photo_reference"`
	Height         int      `json:"height"`
	Width          int      `json:"width"`
	HTMLAttributions []string `json:"html_attributions"`
}

// RestaurantSearchRequest represents search request
type RestaurantSearchRequest struct {
	Query    string  `json:"query" validate:"required,min=2"`
	Location string  `json:"location,omitempty"`
	Radius   int     `json:"radius,omitempty"`
	Type     string  `json:"type,omitempty"`
}

// RestaurantSearchResponse represents search response
type RestaurantSearchResponse struct {
	Restaurants []Restaurant `json:"restaurants"`
	Status      string       `json:"status"`
	NextPageToken string     `json:"next_page_token,omitempty"`
}

// GooglePlacesResponse represents the response from Google Places API
type GooglePlacesResponse struct {
	Results       []Restaurant `json:"results"`
	Status        string       `json:"status"`
	ErrorMessage  string       `json:"error_message,omitempty"`
	NextPageToken string       `json:"next_page_token,omitempty"`
}