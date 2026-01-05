package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"tj-routes/internal/dto"
	"tj-routes/internal/service"
)

// UserPersonalizedHandler handles user personalized data endpoints
type UserPersonalizedHandler struct {
	personalizedService service.UserPersonalizedService
}

// NewUserPersonalizedHandler creates a new UserPersonalizedHandler
func NewUserPersonalizedHandler(personalizedService service.UserPersonalizedService) *UserPersonalizedHandler {
	return &UserPersonalizedHandler{personalizedService: personalizedService}
}

// getUserID extracts user ID from context
func getUserID(c *gin.Context) (uint, bool) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, errors.New("user not authenticated"))
		return 0, false
	}
	return userIDVal.(uint), true
}

// parsePagination parses pagination parameters from query
func parsePagination(c *gin.Context) (page, limit int) {
	page = 1
	limit = 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	return page, limit
}

// =====================
// Favorite Routes
// =====================

// GetFavoriteRoutes returns the user's favorite routes
// @Summary Get favorite routes
// @Description Get all routes favorited by the authenticated user
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} Response{data=dto.UserFavoriteRouteListResponse}
// @Router /api/v1/users/me/personalized/favorites/routes [get]
func (h *UserPersonalizedHandler) GetFavoriteRoutes(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	page, limit := parsePagination(c)

	result, err := h.personalizedService.GetFavoriteRoutes(c.Request.Context(), userID, page, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, result)
}

// AddFavoriteRoute adds a route to user's favorites
// @Summary Add route to favorites
// @Description Add a route to the authenticated user's favorites
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param routeId path int true "Route ID"
// @Success 201 {object} Response{message=string}
// @Router /api/v1/users/me/personalized/favorites/routes/{routeId} [post]
func (h *UserPersonalizedHandler) AddFavoriteRoute(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	routeID, err := strconv.ParseUint(c.Param("routeId"), 10, 64)
	if err != nil {
		BadRequest(c, errors.New("invalid route ID"))
		return
	}

	err = h.personalizedService.AddFavoriteRoute(c.Request.Context(), userID, uint(routeID))
	if err != nil {
		if errors.Is(err, service.ErrFavoriteAlreadyExists) {
			ErrorResponse(c, http.StatusConflict, err)
			return
		}
		InternalServerError(c, err)
		return
	}

	MessageResponse(c, http.StatusCreated, "Route added to favorites")
}

// RemoveFavoriteRoute removes a route from user's favorites
// @Summary Remove route from favorites
// @Description Remove a route from the authenticated user's favorites
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param routeId path int true "Route ID"
// @Success 200 {object} Response{message=string}
// @Router /api/v1/users/me/personalized/favorites/routes/{routeId} [delete]
func (h *UserPersonalizedHandler) RemoveFavoriteRoute(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	routeID, err := strconv.ParseUint(c.Param("routeId"), 10, 64)
	if err != nil {
		BadRequest(c, errors.New("invalid route ID"))
		return
	}

	err = h.personalizedService.RemoveFavoriteRoute(c.Request.Context(), userID, uint(routeID))
	if err != nil {
		InternalServerError(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Route removed from favorites")
}

// IsFavoriteRoute checks if a route is in user's favorites
// @Summary Check if route is favorited
// @Description Check if a route is in the authenticated user's favorites
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param routeId path int true "Route ID"
// @Success 200 {object} Response{data=map[string]bool}
// @Router /api/v1/users/me/personalized/favorites/routes/{routeId}/check [get]
func (h *UserPersonalizedHandler) IsFavoriteRoute(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	routeID, err := strconv.ParseUint(c.Param("routeId"), 10, 64)
	if err != nil {
		BadRequest(c, errors.New("invalid route ID"))
		return
	}

	isFavorite, err := h.personalizedService.IsFavoriteRoute(c.Request.Context(), userID, uint(routeID))
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{"is_favorite": isFavorite})
}

// =====================
// Favorite Stops
// =====================

// GetFavoriteStops returns the user's favorite stops
// @Summary Get favorite stops
// @Description Get all stops favorited by the authenticated user
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} Response{data=dto.UserFavoriteStopListResponse}
// @Router /api/v1/users/me/personalized/favorites/stops [get]
func (h *UserPersonalizedHandler) GetFavoriteStops(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	page, limit := parsePagination(c)

	result, err := h.personalizedService.GetFavoriteStops(c.Request.Context(), userID, page, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, result)
}

// AddFavoriteStop adds a stop to user's favorites
// @Summary Add stop to favorites
// @Description Add a stop to the authenticated user's favorites
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param stopId path int true "Stop ID"
// @Success 201 {object} Response{message=string}
// @Router /api/v1/users/me/personalized/favorites/stops/{stopId} [post]
func (h *UserPersonalizedHandler) AddFavoriteStop(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	stopID, err := strconv.ParseUint(c.Param("stopId"), 10, 64)
	if err != nil {
		BadRequest(c, errors.New("invalid stop ID"))
		return
	}

	err = h.personalizedService.AddFavoriteStop(c.Request.Context(), userID, uint(stopID))
	if err != nil {
		if errors.Is(err, service.ErrFavoriteAlreadyExists) {
			ErrorResponse(c, http.StatusConflict, err)
			return
		}
		InternalServerError(c, err)
		return
	}

	MessageResponse(c, http.StatusCreated, "Stop added to favorites")
}

// RemoveFavoriteStop removes a stop from user's favorites
// @Summary Remove stop from favorites
// @Description Remove a stop from the authenticated user's favorites
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param stopId path int true "Stop ID"
// @Success 200 {object} Response{message=string}
// @Router /api/v1/users/me/personalized/favorites/stops/{stopId} [delete]
func (h *UserPersonalizedHandler) RemoveFavoriteStop(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	stopID, err := strconv.ParseUint(c.Param("stopId"), 10, 64)
	if err != nil {
		BadRequest(c, errors.New("invalid stop ID"))
		return
	}

	err = h.personalizedService.RemoveFavoriteStop(c.Request.Context(), userID, uint(stopID))
	if err != nil {
		InternalServerError(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Stop removed from favorites")
}

// IsFavoriteStop checks if a stop is in user's favorites
// @Summary Check if stop is favorited
// @Description Check if a stop is in the authenticated user's favorites
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param stopId path int true "Stop ID"
// @Success 200 {object} Response{data=map[string]bool}
// @Router /api/v1/users/me/personalized/favorites/stops/{stopId}/check [get]
func (h *UserPersonalizedHandler) IsFavoriteStop(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	stopID, err := strconv.ParseUint(c.Param("stopId"), 10, 64)
	if err != nil {
		BadRequest(c, errors.New("invalid stop ID"))
		return
	}

	isFavorite, err := h.personalizedService.IsFavoriteStop(c.Request.Context(), userID, uint(stopID))
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{"is_favorite": isFavorite})
}

// =====================
// Places
// =====================

// GetPlaces returns all saved places for the user
// @Summary Get saved places
// @Description Get all saved places for the authenticated user
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Success 200 {object} Response{data=dto.UserPlaceListResponse}
// @Router /api/v1/users/me/personalized/places [get]
func (h *UserPersonalizedHandler) GetPlaces(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	result, err := h.personalizedService.GetPlaces(c.Request.Context(), userID)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, result)
}

// CreatePlace creates a new saved place
// @Summary Create saved place
// @Description Create a new saved place for the authenticated user
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param place body dto.CreateUserPlaceRequest true "Place details"
// @Success 201 {object} Response{data=dto.UserPlaceResponse}
// @Router /api/v1/users/me/personalized/places [post]
func (h *UserPersonalizedHandler) CreatePlace(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var req dto.CreateUserPlaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	result, err := h.personalizedService.CreatePlace(c.Request.Context(), userID, &req)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusCreated, result)
}

// GetPlace returns a specific saved place
// @Summary Get saved place
// @Description Get a specific saved place by ID
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param id path int true "Place ID"
// @Success 200 {object} Response{data=dto.UserPlaceResponse}
// @Router /api/v1/users/me/personalized/places/{id} [get]
func (h *UserPersonalizedHandler) GetPlace(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	placeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errors.New("invalid place ID"))
		return
	}

	result, err := h.personalizedService.GetPlaceByID(c.Request.Context(), userID, uint(placeID))
	if err != nil {
		if errors.Is(err, service.ErrPlaceNotFound) || errors.Is(err, service.ErrUnauthorizedPlaceAccess) {
			ErrorResponse(c, http.StatusNotFound, err)
			return
		}
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, result)
}

// UpdatePlace updates a saved place
// @Summary Update saved place
// @Description Update a saved place for the authenticated user
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param id path int true "Place ID"
// @Param place body dto.UpdateUserPlaceRequest true "Place details"
// @Success 200 {object} Response{data=dto.UserPlaceResponse}
// @Router /api/v1/users/me/personalized/places/{id} [put]
func (h *UserPersonalizedHandler) UpdatePlace(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	placeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errors.New("invalid place ID"))
		return
	}

	var req dto.UpdateUserPlaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	result, err := h.personalizedService.UpdatePlace(c.Request.Context(), userID, uint(placeID), &req)
	if err != nil {
		if errors.Is(err, service.ErrPlaceNotFound) || errors.Is(err, service.ErrUnauthorizedPlaceAccess) {
			ErrorResponse(c, http.StatusNotFound, err)
			return
		}
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, result)
}

// DeletePlace deletes a saved place
// @Summary Delete saved place
// @Description Delete a saved place for the authenticated user
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param id path int true "Place ID"
// @Success 200 {object} Response{message=string}
// @Router /api/v1/users/me/personalized/places/{id} [delete]
func (h *UserPersonalizedHandler) DeletePlace(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	placeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errors.New("invalid place ID"))
		return
	}

	err = h.personalizedService.DeletePlace(c.Request.Context(), userID, uint(placeID))
	if err != nil {
		if errors.Is(err, service.ErrPlaceNotFound) || errors.Is(err, service.ErrUnauthorizedPlaceAccess) {
			ErrorResponse(c, http.StatusNotFound, err)
			return
		}
		InternalServerError(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Place deleted successfully")
}

// =====================
// Recent Views
// =====================

// GetRecentRoutes returns recently viewed routes
// @Summary Get recently viewed routes
// @Description Get recently viewed routes for the authenticated user
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} Response{data=dto.UserFavoriteRouteListResponse}
// @Router /api/v1/users/me/personalized/recent/routes [get]
func (h *UserPersonalizedHandler) GetRecentRoutes(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	page, limit := parsePagination(c)

	result, err := h.personalizedService.GetRecentRoutes(c.Request.Context(), userID, page, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, result)
}

// GetRecentStops returns recently viewed stops
// @Summary Get recently viewed stops
// @Description Get recently viewed stops for the authenticated user
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} Response{data=dto.UserFavoriteStopListResponse}
// @Router /api/v1/users/me/personalized/recent/stops [get]
func (h *UserPersonalizedHandler) GetRecentStops(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	page, limit := parsePagination(c)

	result, err := h.personalizedService.GetRecentStops(c.Request.Context(), userID, page, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, result)
}

// GetRecentNavigations returns recent navigation searches
// @Summary Get recent navigation searches
// @Description Get recent navigation searches for the authenticated user
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} Response{data=dto.UserSavedNavigationListResponse}
// @Router /api/v1/users/me/personalized/recent/navigations [get]
func (h *UserPersonalizedHandler) GetRecentNavigations(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	page, limit := parsePagination(c)

	result, err := h.personalizedService.GetRecentNavigations(c.Request.Context(), userID, page, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, result)
}

// RecordRecentNavigation records a navigation search
// @Summary Record navigation search
// @Description Record a navigation search as recent activity
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param navigation body dto.RecordRecentViewRequest true "Navigation details"
// @Success 201 {object} Response{message=string}
// @Router /api/v1/users/me/personalized/recent/navigations [post]
func (h *UserPersonalizedHandler) RecordRecentNavigation(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var req dto.RecordRecentViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	err := h.personalizedService.RecordRecentView(c.Request.Context(), userID, &req)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	MessageResponse(c, http.StatusCreated, "Navigation recorded")
}

// =====================
// Saved Navigations
// =====================

// GetSavedNavigations returns saved navigation pairs
// @Summary Get saved navigation pairs
// @Description Get all saved navigation pairs for the authenticated user
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} Response{data=dto.UserSavedNavigationListResponse}
// @Router /api/v1/users/me/personalized/saved-navigations [get]
func (h *UserPersonalizedHandler) GetSavedNavigations(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	page, limit := parsePagination(c)

	result, err := h.personalizedService.GetSavedNavigations(c.Request.Context(), userID, page, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, result)
}

// CreateSavedNavigation creates a new saved navigation
// @Summary Create saved navigation
// @Description Create a new saved navigation pair for the authenticated user
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param navigation body dto.CreateUserSavedNavigationRequest true "Navigation details"
// @Success 201 {object} Response{data=dto.UserSavedNavigationResponse}
// @Router /api/v1/users/me/personalized/saved-navigations [post]
func (h *UserPersonalizedHandler) CreateSavedNavigation(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var req dto.CreateUserSavedNavigationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	result, err := h.personalizedService.CreateSavedNavigation(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidNavigationPoint) {
			BadRequest(c, err)
			return
		}
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusCreated, result)
}

// UpdateSavedNavigation updates a saved navigation
// @Summary Update saved navigation
// @Description Update a saved navigation pair for the authenticated user
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param id path int true "Navigation ID"
// @Param navigation body dto.UpdateUserSavedNavigationRequest true "Navigation details"
// @Success 200 {object} Response{data=dto.UserSavedNavigationResponse}
// @Router /api/v1/users/me/personalized/saved-navigations/{id} [put]
func (h *UserPersonalizedHandler) UpdateSavedNavigation(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	navID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errors.New("invalid navigation ID"))
		return
	}

	var req dto.UpdateUserSavedNavigationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	result, err := h.personalizedService.UpdateSavedNavigation(c.Request.Context(), userID, uint(navID), &req)
	if err != nil {
		if errors.Is(err, service.ErrSavedNavigationNotFound) || errors.Is(err, service.ErrUnauthorizedAccess) {
			ErrorResponse(c, http.StatusNotFound, err)
			return
		}
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, result)
}

// DeleteSavedNavigation deletes a saved navigation
// @Summary Delete saved navigation
// @Description Delete a saved navigation pair for the authenticated user
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Param id path int true "Navigation ID"
// @Success 200 {object} Response{message=string}
// @Router /api/v1/users/me/personalized/saved-navigations/{id} [delete]
func (h *UserPersonalizedHandler) DeleteSavedNavigation(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	navID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errors.New("invalid navigation ID"))
		return
	}

	err = h.personalizedService.DeleteSavedNavigation(c.Request.Context(), userID, uint(navID))
	if err != nil {
		if errors.Is(err, service.ErrSavedNavigationNotFound) || errors.Is(err, service.ErrUnauthorizedAccess) {
			ErrorResponse(c, http.StatusNotFound, err)
			return
		}
		InternalServerError(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Saved navigation deleted")
}

// =====================
// Analytics
// =====================

// GetAnalytics returns user analytics
// @Summary Get user analytics
// @Description Get aggregated analytics for the authenticated user
// @Tags UserPersonalized
// @Accept json
// @Produce json
// @Success 200 {object} Response{data=dto.UserAnalyticsResponse}
// @Router /api/v1/users/me/personalized/analytics [get]
func (h *UserPersonalizedHandler) GetAnalytics(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	result, err := h.personalizedService.GetUserAnalytics(c.Request.Context(), userID)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, result)
}
