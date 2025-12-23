package mocks

import (
	"tj-routes/internal/models"

	"github.com/stretchr/testify/mock"
)

type MockVehicleRepository struct {
	mock.Mock
}

func NewMockVehicleRepository() *MockVehicleRepository {
	return new(MockVehicleRepository)
}

func (m *MockVehicleRepository) Create(vehicle *models.Vehicle) error {
	args := m.Called(vehicle)
	return args.Error(0)
}

func (m *MockVehicleRepository) FindByID(id uint) (*models.Vehicle, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Vehicle), args.Error(1)
}

func (m *MockVehicleRepository) Update(vehicle *models.Vehicle) error {
	args := m.Called(vehicle)
	return args.Error(0)
}

func (m *MockVehicleRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockVehicleRepository) List(offset, limit int, filters map[string]interface{}) ([]models.Vehicle, int64, error) {
	args := m.Called(offset, limit, filters)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Vehicle), args.Get(1).(int64), args.Error(2)
}

func (m *MockVehicleRepository) FindByVehiclePlate(plate string) (*models.Vehicle, error) {
	args := m.Called(plate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Vehicle), args.Error(1)
}

