package service

import (
	"encoding/json"
	"fmt"

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
}

func NewStopService(stopRepo repository.StopRepository) StopService {
	return &stopService{
		stopRepo: stopRepo,
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
	return s.stopRepo.Create(stop)
}

func (s *stopService) GetStopByID(id uint) (*models.Stop, error) {
	return s.stopRepo.FindByID(id)
}

func (s *stopService) UpdateStop(stop *models.Stop) error {
	if stop.Facilities != nil {
		// Validate JSON if provided
		var test interface{}
		if err := json.Unmarshal([]byte(*stop.Facilities), &test); err != nil {
			return fmt.Errorf("invalid facilities JSON: %w", err)
		}
	}
	return s.stopRepo.Update(stop)
}

func (s *stopService) DeleteStop(id uint) error {
	return s.stopRepo.Delete(id)
}

func (s *stopService) ListStops(offset, limit int, filters map[string]interface{}) ([]models.Stop, int64, error) {
	return s.stopRepo.List(offset, limit, filters)
}

