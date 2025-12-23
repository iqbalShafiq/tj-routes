package mocks

import (
	"tj-routes/internal/models"

	"github.com/stretchr/testify/mock"
)

type MockStopRepository struct {
	mock.Mock
}

func NewMockStopRepository() *MockStopRepository {
	return new(MockStopRepository)
}

func (m *MockStopRepository) Create(stop *models.Stop) error {
	args := m.Called(stop)
	return args.Error(0)
}

func (m *MockStopRepository) FindByID(id uint) (*models.Stop, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Stop), args.Error(1)
}

func (m *MockStopRepository) Update(stop *models.Stop) error {
	args := m.Called(stop)
	return args.Error(0)
}

func (m *MockStopRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStopRepository) List(offset, limit int, filters map[string]interface{}) ([]models.Stop, int64, error) {
	args := m.Called(offset, limit, filters)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Stop), args.Get(1).(int64), args.Error(2)
}

func (m *MockStopRepository) FindByLatitudeAndLongitude(lat, lng float64) (*models.Stop, error) {
	args := m.Called(lat, lng)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Stop), args.Error(1)
}

