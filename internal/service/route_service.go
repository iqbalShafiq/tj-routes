package service

import (
	"encoding/json"
	"fmt"

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
}

func NewRouteService(
	routeRepo repository.RouteRepository,
	routeStopRepo repository.RouteStopRepository,
	stopRepo repository.StopRepository,
	routeChangeRepo repository.RouteChangeRepository,
) RouteService {
	return &routeService{
		routeRepo:      routeRepo,
		routeStopRepo:  routeStopRepo,
		stopRepo:       stopRepo,
		routeChangeRepo: routeChangeRepo,
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

	return nil
}

func (s *routeService) GetRouteByID(id uint) (*models.Route, error) {
	return s.routeRepo.FindByID(id)
}

func (s *routeService) UpdateRoute(route *models.Route) error {
	return s.routeRepo.Update(route)
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

	return nil
}

func (s *routeService) DeleteRoute(id uint) error {
	// Delete route stops first
	if err := s.routeStopRepo.DeleteByRouteID(id); err != nil {
		return err
	}
	return s.routeRepo.Delete(id)
}

func (s *routeService) ListRoutes(offset, limit int, filters map[string]interface{}) ([]models.Route, int64, error) {
	return s.routeRepo.List(offset, limit, filters)
}

func stringPtr(s string) *string {
	return &s
}

