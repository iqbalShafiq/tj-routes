package service

import (
	"context"
	"time"

	"tj-routes/internal/cache"
	"tj-routes/internal/config"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type VehicleService interface {
	CreateVehicle(vehicle *models.Vehicle) error
	GetVehicleByID(id uint) (*models.Vehicle, error)
	UpdateVehicle(vehicle *models.Vehicle) error
	DeleteVehicle(id uint) error
	ListVehicles(offset, limit int, filters map[string]interface{}) ([]models.Vehicle, int64, error)
}

type vehicleService struct {
	vehicleRepo repository.VehicleRepository
	routeRepo   repository.RouteRepository
	cache       cache.Cache
	config      *config.Config
}

func NewVehicleService(vehicleRepo repository.VehicleRepository, routeRepo repository.RouteRepository, cacheInstance cache.Cache, cfg *config.Config) VehicleService {
	return &vehicleService{
		vehicleRepo: vehicleRepo,
		routeRepo:   routeRepo,
		cache:       cacheInstance,
		config:      cfg,
	}
}

func (s *vehicleService) CreateVehicle(vehicle *models.Vehicle) error {
	// Validate route exists
	_, err := s.routeRepo.FindByID(vehicle.RouteID)
	if err != nil {
		return err
	}
	if err := s.vehicleRepo.Create(vehicle); err != nil {
		return err
	}

	// Invalidate cache
	ctx := context.Background()
	s.cache.InvalidatePattern(ctx, cache.VehiclePattern())
	s.cache.Delete(ctx, cache.RouteKey(vehicle.RouteID))

	return nil
}

func (s *vehicleService) GetVehicleByID(id uint) (*models.Vehicle, error) {
	ctx := context.Background()
	key := cache.VehicleKey(id)

	// Try to get from cache
	var vehicle models.Vehicle
	if err := s.cache.Get(ctx, key, &vehicle); err == nil {
		return &vehicle, nil
	}

	// Cache miss, get from database
	vehiclePtr, err := s.vehicleRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Store in cache
	ttl := time.Duration(s.config.Cache.VehicleTTL) * time.Minute
	s.cache.Set(ctx, key, *vehiclePtr, ttl)

	return vehiclePtr, nil
}

func (s *vehicleService) UpdateVehicle(vehicle *models.Vehicle) error {
	// Validate route exists if being changed
	if vehicle.RouteID > 0 {
		_, err := s.routeRepo.FindByID(vehicle.RouteID)
		if err != nil {
			return err
		}
	}
	if err := s.vehicleRepo.Update(vehicle); err != nil {
		return err
	}

	// Invalidate cache
	ctx := context.Background()
	s.cache.Delete(ctx, cache.VehicleKey(vehicle.ID))
	s.cache.InvalidatePattern(ctx, cache.VehiclePattern())
	if vehicle.RouteID > 0 {
		s.cache.Delete(ctx, cache.RouteKey(vehicle.RouteID))
	}

	return nil
}

func (s *vehicleService) DeleteVehicle(id uint) error {
	// Get vehicle to know route ID for cache invalidation
	vehicle, err := s.vehicleRepo.FindByID(id)
	if err != nil {
		return err
	}

	if err := s.vehicleRepo.Delete(id); err != nil {
		return err
	}

	// Invalidate cache
	ctx := context.Background()
	s.cache.Delete(ctx, cache.VehicleKey(id))
	s.cache.InvalidatePattern(ctx, cache.VehiclePattern())
	if vehicle != nil && vehicle.RouteID > 0 {
		s.cache.Delete(ctx, cache.RouteKey(vehicle.RouteID))
	}

	return nil
}

func (s *vehicleService) ListVehicles(offset, limit int, filters map[string]interface{}) ([]models.Vehicle, int64, error) {
	ctx := context.Background()

	// Build cache key from filters
	page := (offset / limit) + 1
	if limit == 0 {
		limit = 10
	}
	status := ""
	routeID := uint(0)
	if s, ok := filters["status"].(models.Status); ok {
		status = string(s)
	}
	if rid, ok := filters["route_id"].(uint); ok {
		routeID = rid
	}
	key := cache.VehicleListKey(page, limit, status, routeID)

	// Try to get from cache
	type cachedResult struct {
		Vehicles []models.Vehicle `json:"vehicles"`
		Total    int64            `json:"total"`
	}
	var cached cachedResult
	if err := s.cache.Get(ctx, key, &cached); err == nil {
		return cached.Vehicles, cached.Total, nil
	}

	// Cache miss, get from database
	vehicles, total, err := s.vehicleRepo.List(offset, limit, filters)
	if err != nil {
		return nil, 0, err
	}

	// Store in cache
	ttl := time.Duration(s.config.Cache.VehicleTTL) * time.Minute
	s.cache.Set(ctx, key, cachedResult{Vehicles: vehicles, Total: total}, ttl)

	return vehicles, total, nil
}

