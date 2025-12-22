package mocks

import (
	"tj-routes/internal/models"

	"github.com/stretchr/testify/mock"
)

type MockRouteStopRepository struct {
	mock.Mock
}

func NewMockRouteStopRepository() *MockRouteStopRepository {
	return new(MockRouteStopRepository)
}

func (m *MockRouteStopRepository) Create(routeStop *models.RouteStop) error {
	args := m.Called(routeStop)
	return args.Error(0)
}

func (m *MockRouteStopRepository) FindByRouteID(routeID uint) ([]models.RouteStop, error) {
	args := m.Called(routeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RouteStop), args.Error(1)
}

func (m *MockRouteStopRepository) FindByID(id uint) (*models.RouteStop, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RouteStop), args.Error(1)
}

func (m *MockRouteStopRepository) Update(routeStop *models.RouteStop) error {
	args := m.Called(routeStop)
	return args.Error(0)
}

func (m *MockRouteStopRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRouteStopRepository) DeleteByRouteID(routeID uint) error {
	args := m.Called(routeID)
	return args.Error(0)
}

func (m *MockRouteStopRepository) DeleteByRouteAndStop(routeID, stopID uint) error {
	args := m.Called(routeID, stopID)
	return args.Error(0)
}

