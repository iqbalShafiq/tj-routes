package service

import (
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
}

func NewVehicleService(vehicleRepo repository.VehicleRepository, routeRepo repository.RouteRepository) VehicleService {
	return &vehicleService{
		vehicleRepo: vehicleRepo,
		routeRepo:   routeRepo,
	}
}

func (s *vehicleService) CreateVehicle(vehicle *models.Vehicle) error {
	// Validate route exists
	_, err := s.routeRepo.FindByID(vehicle.RouteID)
	if err != nil {
		return err
	}
	return s.vehicleRepo.Create(vehicle)
}

func (s *vehicleService) GetVehicleByID(id uint) (*models.Vehicle, error) {
	return s.vehicleRepo.FindByID(id)
}

func (s *vehicleService) UpdateVehicle(vehicle *models.Vehicle) error {
	// Validate route exists if being changed
	if vehicle.RouteID > 0 {
		_, err := s.routeRepo.FindByID(vehicle.RouteID)
		if err != nil {
			return err
		}
	}
	return s.vehicleRepo.Update(vehicle)
}

func (s *vehicleService) DeleteVehicle(id uint) error {
	return s.vehicleRepo.Delete(id)
}

func (s *vehicleService) ListVehicles(offset, limit int, filters map[string]interface{}) ([]models.Vehicle, int64, error) {
	return s.vehicleRepo.List(offset, limit, filters)
}

