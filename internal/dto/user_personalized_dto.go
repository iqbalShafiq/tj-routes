package dto

import (
	"time"

	"tj-routes/internal/models"
)

// =====================
// Request DTOs
// =====================

// CreateUserPlaceRequest is the request body for creating a saved place
type CreateUserPlaceRequest struct {
	PlaceType models.PlaceType `json:"place_type" binding:"required"`
	Name      string           `json:"name" binding:"required"`
	Latitude  float64          `json:"latitude" binding:"required"`
	Longitude float64          `json:"longitude" binding:"required"`
	Address   *string          `json:"address,omitempty"`
	Notes     *string          `json:"notes,omitempty"`
}

// UpdateUserPlaceRequest is the request body for updating a saved place
type UpdateUserPlaceRequest struct {
	PlaceType *models.PlaceType `json:"place_type,omitempty"`
	Name      *string           `json:"name,omitempty"`
	Latitude  *float64          `json:"latitude,omitempty"`
	Longitude *float64          `json:"longitude,omitempty"`
	Address   *string           `json:"address,omitempty"`
	Notes     *string           `json:"notes,omitempty"`
	IsDefault *bool             `json:"is_default,omitempty"`
}

// CreateUserSavedNavigationRequest is the request body for creating a saved navigation
type CreateUserSavedNavigationRequest struct {
	Name          *string  `json:"name,omitempty"`
	FromPlaceType *string  `json:"from_place_type,omitempty"`
	FromPlaceID   *uint    `json:"from_place_id,omitempty"`
	FromStopID    *uint    `json:"from_stop_id,omitempty"`
	ToPlaceType   *string  `json:"to_place_type,omitempty"`
	ToPlaceID     *uint    `json:"to_place_id,omitempty"`
	ToStopID      *uint    `json:"to_stop_id,omitempty"`
}

// UpdateUserSavedNavigationRequest is the request body for updating a saved navigation
type UpdateUserSavedNavigationRequest struct {
	Name          *string `json:"name,omitempty"`
	FromPlaceType *string `json:"from_place_type,omitempty"`
	FromPlaceID   *uint   `json:"from_place_id,omitempty"`
	FromStopID    *uint   `json:"from_stop_id,omitempty"`
	ToPlaceType   *string `json:"to_place_type,omitempty"`
	ToPlaceID     *uint   `json:"to_place_id,omitempty"`
	ToStopID      *uint   `json:"to_stop_id,omitempty"`
}

// RecordRecentViewRequest is the request body for recording a recent view
type RecordRecentViewRequest struct {
	ViewType   models.RecentViewType `json:"view_type" binding:"required"`
	RouteID    *uint                 `json:"route_id,omitempty"`
	FromStopID *uint                 `json:"from_stop_id,omitempty"`
	ToStopID   *uint                 `json:"to_stop_id,omitempty"`
}

// =====================
// Response DTOs
// =====================

// UserPlaceResponse is the response for a saved place
type UserPlaceResponse struct {
	ID        uint             `json:"id"`
	PlaceType models.PlaceType `json:"place_type"`
	Name      string           `json:"name"`
	Latitude  float64          `json:"latitude"`
	Longitude float64          `json:"longitude"`
	Address   string           `json:"address,omitempty"`
	Notes     string           `json:"notes,omitempty"`
	IsDefault bool             `json:"is_default"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// UserFavoriteRouteResponse is the response for a favorite route
type UserFavoriteRouteResponse struct {
	ID        uint                 `json:"id"`
	Route     models.Route         `json:"route"`
	CreatedAt time.Time            `json:"created_at"`
}

// UserFavoriteStopResponse is the response for a favorite stop
type UserFavoriteStopResponse struct {
	ID        uint         `json:"id"`
	Stop      models.Stop  `json:"stop"`
	CreatedAt time.Time    `json:"created_at"`
}

// NavigationPoint represents a navigation point in a saved navigation
type NavigationPoint struct {
	PlaceType string  `json:"place_type,omitempty"`
	PlaceID   *uint   `json:"place_id,omitempty"`
	PlaceName string  `json:"place_name,omitempty"`
	StopID    *uint   `json:"stop_id,omitempty"`
	StopName  string  `json:"stop_name,omitempty"`
}

// UserSavedNavigationResponse is the response for a saved navigation
type UserSavedNavigationResponse struct {
	ID        uint            `json:"id"`
	Name      string          `json:"name,omitempty"`
	From      NavigationPoint `json:"from"`
	To        NavigationPoint `json:"to"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// RouteStats represents statistics for a route
type RouteStats struct {
	RouteID     uint   `json:"route_id"`
	RouteNumber string `json:"route_number"`
	RouteName   string `json:"route_name"`
	UsageCount  int64  `json:"usage_count"`
}

// StopStats represents statistics for a stop
type StopStats struct {
	StopID     uint   `json:"stop_id"`
	StopName   string `json:"stop_name"`
	VisitCount int64  `json:"visit_count"`
}

// UserAnalyticsResponse is the response for user analytics
type UserAnalyticsResponse struct {
	MostFrequentRoute     *RouteStats `json:"most_frequent_route,omitempty"`
	MostFrequentStop      *StopStats  `json:"most_frequent_stop,omitempty"`
	TotalCheckIns         int64       `json:"total_check_ins"`
	TotalRoutesTraveled   int64       `json:"total_routes_traveled"`
	TotalUniqueRoutes     int64       `json:"total_unique_routes"`
	TotalDurationSeconds  int64       `json:"total_duration_seconds"`
	FavoritePlacesCount   int         `json:"favorite_places_count"`
	FavoriteRoutesCount   int         `json:"favorite_routes_count"`
	FavoriteStopsCount    int         `json:"favorite_stops_count"`
	SavedNavigationsCount int         `json:"saved_navigations_count"`
}

// =====================
// Pagination Responses
// =====================

// UserFavoriteRouteListResponse is the paginated response for favorite routes
type UserFavoriteRouteListResponse struct {
	Data  []UserFavoriteRouteResponse `json:"data"`
	Total int64                       `json:"total"`
	Page  int                         `json:"page"`
	Limit int                         `json:"limit"`
}

// UserFavoriteStopListResponse is the paginated response for favorite stops
type UserFavoriteStopListResponse struct {
	Data  []UserFavoriteStopResponse `json:"data"`
	Total int64                      `json:"total"`
	Page  int                        `json:"page"`
	Limit int                        `json:"limit"`
}

// UserPlaceListResponse is the response for saved places
type UserPlaceListResponse struct {
	Data  []UserPlaceResponse `json:"data"`
	Total int64               `json:"total"`
}

// UserSavedNavigationListResponse is the paginated response for saved navigations
type UserSavedNavigationListResponse struct {
	Data  []UserSavedNavigationResponse `json:"data"`
	Total int64                         `json:"total"`
	Page  int                           `json:"page"`
	Limit int                           `json:"limit"`
}
