package mocks

import (
	"tj-routes/internal/models"

	"github.com/stretchr/testify/mock"
)

type MockRouteRepository struct {
	mock.Mock
}

func NewMockRouteRepository() *MockRouteRepository {
	return new(MockRouteRepository)
}

func (m *MockRouteRepository) Create(route *models.Route) error {
	args := m.Called(route)
	return args.Error(0)
}

func (m *MockRouteRepository) FindByID(id uint) (*models.Route, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Route), args.Error(1)
}

func (m *MockRouteRepository) Update(route *models.Route) error {
	args := m.Called(route)
	return args.Error(0)
}

func (m *MockRouteRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRouteRepository) List(offset, limit int, filters map[string]interface{}) ([]models.Route, int64, error) {
	args := m.Called(offset, limit, filters)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Route), args.Get(1).(int64), args.Error(2)
}

