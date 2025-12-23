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

type StopService interface {
	CreateStop(stop *models.Stop) error
	GetStopByID(id uint) (*models.Stop, error)
	UpdateStop(stop *models.Stop) error
	DeleteStop(id uint) error
	ListStops(offset, limit int, filters map[string]interface{}) ([]models.Stop, int64, error)
}

type stopService struct {
	stopRepo repository.StopRepository
	cache    cache.Cache
	config   *config.Config
}

func NewStopService(stopRepo repository.StopRepository, cacheInstance cache.Cache, cfg *config.Config) StopService {
	return &stopService{
		stopRepo: stopRepo,
		cache:    cacheInstance,
		config:   cfg,
	}
}

func (s *stopService) CreateStop(stop *models.Stop) error {
	if stop.Facilities != nil {
		// Validate JSON if provided
		var test interface{}
		if err := json.Unmarshal([]byte(*stop.Facilities), &test); err != nil {
			return fmt.Errorf("invalid facilities JSON: %w", err)
		}
	}
	if err := s.stopRepo.Create(stop); err != nil {
		return err
	}

	// Invalidate cache (stops affect routes too)
	ctx := context.Background()
	s.cache.InvalidatePattern(ctx, cache.StopPattern())
	s.cache.InvalidatePattern(ctx, cache.RoutePattern())

	return nil
}

func (s *stopService) GetStopByID(id uint) (*models.Stop, error) {
	ctx := context.Background()
	key := cache.StopKey(id)

	// Try to get from cache
	var stop models.Stop
	if err := s.cache.Get(ctx, key, &stop); err == nil {
		return &stop, nil
	}

	// Cache miss, get from database
	stopPtr, err := s.stopRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Store in cache
	ttl := time.Duration(s.config.Cache.StopTTL) * time.Minute
	s.cache.Set(ctx, key, *stopPtr, ttl)

	return stopPtr, nil
}

func (s *stopService) UpdateStop(stop *models.Stop) error {
	if stop.Facilities != nil {
		// Validate JSON if provided
		var test interface{}
		if err := json.Unmarshal([]byte(*stop.Facilities), &test); err != nil {
			return fmt.Errorf("invalid facilities JSON: %w", err)
		}
	}
	if err := s.stopRepo.Update(stop); err != nil {
		return err
	}

	// Invalidate cache (stops affect routes too)
	ctx := context.Background()
	s.cache.Delete(ctx, cache.StopKey(stop.ID))
	s.cache.InvalidatePattern(ctx, cache.StopPattern())
	s.cache.InvalidatePattern(ctx, cache.RoutePattern())

	return nil
}

func (s *stopService) DeleteStop(id uint) error {
	if err := s.stopRepo.Delete(id); err != nil {
		return err
	}

	// Invalidate cache (stops affect routes too)
	ctx := context.Background()
	s.cache.Delete(ctx, cache.StopKey(id))
	s.cache.InvalidatePattern(ctx, cache.StopPattern())
	s.cache.InvalidatePattern(ctx, cache.RoutePattern())

	return nil
}

func (s *stopService) ListStops(offset, limit int, filters map[string]interface{}) ([]models.Stop, int64, error) {
	ctx := context.Background()

	// Build cache key from filters
	page := (offset / limit) + 1
	if limit == 0 {
		limit = 10
	}
	status := ""
	stopType := ""
	city := ""
	if s, ok := filters["status"].(models.Status); ok {
		status = string(s)
	}
	if st, ok := filters["type"].(models.StopType); ok {
		stopType = string(st)
	}
	if c, ok := filters["city"].(string); ok {
		city = c
	}
	key := cache.StopListKey(page, limit, status, stopType, city)

	// Try to get from cache
	type cachedResult struct {
		Stops []models.Stop `json:"stops"`
		Total int64         `json:"total"`
	}
	var cached cachedResult
	if err := s.cache.Get(ctx, key, &cached); err == nil {
		return cached.Stops, cached.Total, nil
	}

	// Cache miss, get from database
	stops, total, err := s.stopRepo.List(offset, limit, filters)
	if err != nil {
		return nil, 0, err
	}

	// Store in cache
	ttl := time.Duration(s.config.Cache.StopTTL) * time.Minute
	s.cache.Set(ctx, key, cachedResult{Stops: stops, Total: total}, ttl)

	return stops, total, nil
}

