package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"tj-routes/internal/cache"
	"tj-routes/internal/config"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type RouteService interface {
	CreateRoute(route *models.Route, stopIDs []uint, userID uint) error
	GetRouteByID(id uint) (*models.Route, error)
	UpdateRoute(route *models.Route) error
	UpdateRouteStops(routeID uint, stopIDs []uint, userID uint) error
	DeleteRoute(id uint) error
	ListRoutes(offset, limit int, filters map[string]interface{}) ([]models.Route, int64, error)
}

type routeService struct {
	routeRepo      repository.RouteRepository
	routeStopRepo  repository.RouteStopRepository
	stopRepo       repository.StopRepository
	routeChangeRepo repository.RouteChangeRepository
	cache          cache.Cache
	config         *config.Config
}

func NewRouteService(
	routeRepo repository.RouteRepository,
	routeStopRepo repository.RouteStopRepository,
	stopRepo repository.StopRepository,
	routeChangeRepo repository.RouteChangeRepository,
	cacheInstance cache.Cache,
	cfg *config.Config,
) RouteService {
	return &routeService{
		routeRepo:      routeRepo,
		routeStopRepo:  routeStopRepo,
		stopRepo:       stopRepo,
		routeChangeRepo: routeChangeRepo,
		cache:          cacheInstance,
		config:         cfg,
	}
}

func (s *routeService) CreateRoute(route *models.Route, stopIDs []uint, userID uint) error {
	// Validate stops exist
	for _, stopID := range stopIDs {
		_, err := s.stopRepo.FindByID(stopID)
		if err != nil {
			return fmt.Errorf("stop with ID %d not found", stopID)
		}
	}

	// Create route
	if err := s.routeRepo.Create(route); err != nil {
		return err
	}

	// Create route stops with sequence
	for i, stopID := range stopIDs {
		routeStop := &models.RouteStop{
			RouteID:       route.ID,
			StopID:        stopID,
			SequenceOrder: i + 1,
			IsOrigin:      i == 0,
			IsDestination: i == len(stopIDs)-1,
		}
		if err := s.routeStopRepo.Create(routeStop); err != nil {
			return fmt.Errorf("failed to create route stop: %w", err)
		}
	}

	// Log route creation
	newData, _ := json.Marshal(map[string]interface{}{
		"route_number": route.RouteNumber,
		"stops":        stopIDs,
	})

	routeChange := &models.RouteChange{
		RouteID:    route.ID,
		ChangedBy:  userID,
		ChangeType: models.ChangeTypeRouteCreated,
		OldData:    nil,
		NewData:    stringPtr(string(newData)),
		Reason:     "Route created",
	}
	s.routeChangeRepo.Create(routeChange)

	// Invalidate cache
	ctx := context.Background()
	s.cache.InvalidatePattern(ctx, cache.RoutePattern())

	return nil
}

func (s *routeService) GetRouteByID(id uint) (*models.Route, error) {
	ctx := context.Background()
	key := cache.RouteKey(id)

	// Try to get from cache
	var route models.Route
	if err := s.cache.Get(ctx, key, &route); err == nil {
		return &route, nil
	}

	// Cache miss, get from database
	routePtr, err := s.routeRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Store in cache
	ttl := time.Duration(s.config.Cache.RouteTTL) * time.Minute
	s.cache.Set(ctx, key, *routePtr, ttl)

	return routePtr, nil
}

func (s *routeService) UpdateRoute(route *models.Route) error {
	if err := s.routeRepo.Update(route); err != nil {
		return err
	}

	// Invalidate cache
	ctx := context.Background()
	s.cache.Delete(ctx, cache.RouteKey(route.ID))
	s.cache.InvalidatePattern(ctx, cache.RoutePattern())

	return nil
}

func (s *routeService) UpdateRouteStops(routeID uint, stopIDs []uint, userID uint) error {
	// Validate route exists
	_, err := s.routeRepo.FindByID(routeID)
	if err != nil {
		return err
	}

	// Get existing route stops
	existingRouteStops, err := s.routeStopRepo.FindByRouteID(routeID)
	if err != nil {
		return err
	}

	// Store old data for audit
	oldStopIDs := make([]uint, len(existingRouteStops))
	for i, rs := range existingRouteStops {
		oldStopIDs[i] = rs.StopID
	}

	// Validate new stops exist
	for _, stopID := range stopIDs {
		_, err := s.stopRepo.FindByID(stopID)
		if err != nil {
			return fmt.Errorf("stop with ID %d not found", stopID)
		}
	}

	// Delete existing route stops
	if err := s.routeStopRepo.DeleteByRouteID(routeID); err != nil {
		return err
	}

	// Create new route stops
	for i, stopID := range stopIDs {
		routeStop := &models.RouteStop{
			RouteID:       routeID,
			StopID:        stopID,
			SequenceOrder: i + 1,
			IsOrigin:      i == 0,
			IsDestination: i == len(stopIDs)-1,
		}
		if err := s.routeStopRepo.Create(routeStop); err != nil {
			return fmt.Errorf("failed to create route stop: %w", err)
		}
	}

	// Log route change
	oldData, _ := json.Marshal(map[string]interface{}{
		"stops": oldStopIDs,
	})
	newData, _ := json.Marshal(map[string]interface{}{
		"stops": stopIDs,
	})

	routeChange := &models.RouteChange{
		RouteID:    routeID,
		ChangedBy:  userID,
		ChangeType: models.ChangeTypeStopOrderChanged,
		OldData:    stringPtr(string(oldData)),
		NewData:    stringPtr(string(newData)),
		Reason:     "Route stops updated",
	}
	s.routeChangeRepo.Create(routeChange)

	// Invalidate cache
	ctx := context.Background()
	s.cache.Delete(ctx, cache.RouteKey(routeID))
	s.cache.InvalidatePattern(ctx, cache.RoutePattern())

	return nil
}

func (s *routeService) DeleteRoute(id uint) error {
	// Delete route stops first
	if err := s.routeStopRepo.DeleteByRouteID(id); err != nil {
		return err
	}
	if err := s.routeRepo.Delete(id); err != nil {
		return err
	}

	// Invalidate cache
	ctx := context.Background()
	s.cache.Delete(ctx, cache.RouteKey(id))
	s.cache.InvalidatePattern(ctx, cache.RoutePattern())

	return nil
}

func (s *routeService) ListRoutes(offset, limit int, filters map[string]interface{}) ([]models.Route, int64, error) {
	ctx := context.Background()

	// Build cache key from filters
	page := (offset / limit) + 1
	if limit == 0 {
		limit = 10
	}
	status := ""
	routeNumber := ""
	search := ""
	if s, ok := filters["status"].(models.Status); ok {
		status = string(s)
	}
	if rn, ok := filters["route_number"].(string); ok {
		routeNumber = rn
	}
	if s, ok := filters["search"].(string); ok {
		search = s
	}
	key := cache.RouteListKey(page, limit, status, routeNumber, search)

	// Try to get from cache
	type cachedResult struct {
		Routes []models.Route `json:"routes"`
		Total  int64          `json:"total"`
	}
	var cached cachedResult
	if err := s.cache.Get(ctx, key, &cached); err == nil {
		return cached.Routes, cached.Total, nil
	}

	// Cache miss, get from database
	routes, total, err := s.routeRepo.List(offset, limit, filters)
	if err != nil {
		return nil, 0, err
	}

	// Store in cache
	ttl := time.Duration(s.config.Cache.RouteTTL) * time.Minute
	s.cache.Set(ctx, key, cachedResult{Routes: routes, Total: total}, ttl)

	return routes, total, nil
}

func stringPtr(s string) *string {
	return &s
}

