package mocks

import (
	"tj-routes/internal/models"

	"github.com/stretchr/testify/mock"
)

type MockRouteChangeRepository struct {
	mock.Mock
}

func NewMockRouteChangeRepository() *MockRouteChangeRepository {
	return new(MockRouteChangeRepository)
}

func (m *MockRouteChangeRepository) Create(routeChange *models.RouteChange) error {
	args := m.Called(routeChange)
	return args.Error(0)
}

func (m *MockRouteChangeRepository) FindByRouteID(routeID uint) ([]models.RouteChange, error) {
	args := m.Called(routeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RouteChange), args.Error(1)
}

func (m *MockRouteChangeRepository) FindByID(id uint) (*models.RouteChange, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RouteChange), args.Error(1)
}

