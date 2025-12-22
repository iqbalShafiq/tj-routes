package service

import (
	"testing"

	"tj-routes/internal/models"
	"tj-routes/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestRouteService_CreateRoute(t *testing.T) {
	tests := []struct {
		name       string
		route      *models.Route
		stopIDs    []uint
		userID     uint
		setupMocks func(*mocks.MockRouteRepository, *mocks.MockRouteStopRepository, *mocks.MockStopRepository, *mocks.MockRouteChangeRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful creation with stops",
			route: &models.Route{
				RouteNumber: "R9",
				Name:        "Route 9",
				Description: "Test Route",
				Status:      models.StatusActive,
			},
			stopIDs: []uint{1, 2, 3},
			userID:  1,
			setupMocks: func(rm *mocks.MockRouteRepository, rsm *mocks.MockRouteStopRepository, sm *mocks.MockStopRepository, rcm *mocks.MockRouteChangeRepository) {
				// Validate stops exist
				sm.On("FindByID", uint(1)).Return(&models.Stop{ID: 1}, nil)
				sm.On("FindByID", uint(2)).Return(&models.Stop{ID: 2}, nil)
				sm.On("FindByID", uint(3)).Return(&models.Stop{ID: 3}, nil)
				// Create route
				rm.On("Create", mock.AnythingOfType("*models.Route")).Return(nil).Run(func(args mock.Arguments) {
					route := args.Get(0).(*models.Route)
					route.ID = 1 // Simulate DB assigning ID
				})
				// Create route stops
				rsm.On("Create", mock.AnythingOfType("*models.RouteStop")).Return(nil).Times(3)
				// Log route change
				rcm.On("Create", mock.AnythingOfType("*models.RouteChange")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "stop not found",
			route: &models.Route{
				RouteNumber: "R9",
				Name:        "Route 9",
				Status:      models.StatusActive,
			},
			stopIDs: []uint{1, 999},
			userID:  1,
			setupMocks: func(rm *mocks.MockRouteRepository, rsm *mocks.MockRouteStopRepository, sm *mocks.MockStopRepository, rcm *mocks.MockRouteChangeRepository) {
				sm.On("FindByID", uint(1)).Return(&models.Stop{ID: 1}, nil)
				sm.On("FindByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "stop with ID 999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRouteRepo := mocks.NewMockRouteRepository()
			mockRouteStopRepo := mocks.NewMockRouteStopRepository()
			mockStopRepo := mocks.NewMockStopRepository()
			mockRouteChangeRepo := mocks.NewMockRouteChangeRepository()
			tt.setupMocks(mockRouteRepo, mockRouteStopRepo, mockStopRepo, mockRouteChangeRepo)

			service := NewRouteService(mockRouteRepo, mockRouteStopRepo, mockStopRepo, mockRouteChangeRepo)
			err := service.CreateRoute(tt.route, tt.stopIDs, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				// Verify route stops were created with correct sequence
				mockRouteStopRepo.AssertNumberOfCalls(t, "Create", len(tt.stopIDs))
			}

			mockRouteRepo.AssertExpectations(t)
			mockRouteStopRepo.AssertExpectations(t)
			mockStopRepo.AssertExpectations(t)
			mockRouteChangeRepo.AssertExpectations(t)
		})
	}
}

func TestRouteService_GetRouteByID(t *testing.T) {
	mockRepo := mocks.NewMockRouteRepository()
	expectedRoute := &models.Route{
		ID:          1,
		RouteNumber: "R9",
		Name:        "Route 9",
		Status:      models.StatusActive,
	}

	mockRepo.On("FindByID", uint(1)).Return(expectedRoute, nil)

	service := NewRouteService(mockRepo, mocks.NewMockRouteStopRepository(), mocks.NewMockStopRepository(), mocks.NewMockRouteChangeRepository())
	route, err := service.GetRouteByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, route)
	assert.Equal(t, expectedRoute.ID, route.ID)
	assert.Equal(t, expectedRoute.RouteNumber, route.RouteNumber)

	mockRepo.AssertExpectations(t)
}

func TestRouteService_UpdateRoute(t *testing.T) {
	mockRepo := mocks.NewMockRouteRepository()
	route := &models.Route{
		ID:          1,
		RouteNumber: "R9",
		Name:        "Updated Route 9",
		Status:      models.StatusActive,
	}

	mockRepo.On("Update", route).Return(nil)

	service := NewRouteService(mockRepo, mocks.NewMockRouteStopRepository(), mocks.NewMockStopRepository(), mocks.NewMockRouteChangeRepository())
	err := service.UpdateRoute(route)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRouteService_UpdateRouteStops(t *testing.T) {
	tests := []struct {
		name       string
		routeID    uint
		stopIDs    []uint
		userID     uint
		setupMocks func(*mocks.MockRouteRepository, *mocks.MockRouteStopRepository, *mocks.MockStopRepository, *mocks.MockRouteChangeRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:    "successful update",
			routeID: 1,
			stopIDs: []uint{1, 2, 3, 4},
			userID:  1,
			setupMocks: func(rm *mocks.MockRouteRepository, rsm *mocks.MockRouteStopRepository, sm *mocks.MockStopRepository, rcm *mocks.MockRouteChangeRepository) {
				// Validate route exists
				rm.On("FindByID", uint(1)).Return(&models.Route{ID: 1}, nil)
				// Get existing route stops
				existingStops := []models.RouteStop{
					{RouteID: 1, StopID: 1, SequenceOrder: 1},
					{RouteID: 1, StopID: 2, SequenceOrder: 2},
				}
				rsm.On("FindByRouteID", uint(1)).Return(existingStops, nil)
				// Validate new stops exist
				sm.On("FindByID", uint(1)).Return(&models.Stop{ID: 1}, nil)
				sm.On("FindByID", uint(2)).Return(&models.Stop{ID: 2}, nil)
				sm.On("FindByID", uint(3)).Return(&models.Stop{ID: 3}, nil)
				sm.On("FindByID", uint(4)).Return(&models.Stop{ID: 4}, nil)
				// Delete existing route stops
				rsm.On("DeleteByRouteID", uint(1)).Return(nil)
				// Create new route stops
				rsm.On("Create", mock.AnythingOfType("*models.RouteStop")).Return(nil).Times(4)
				// Log route change
				rcm.On("Create", mock.AnythingOfType("*models.RouteChange")).Return(nil)
			},
			expectError: false,
		},
		{
			name:    "route not found",
			routeID: 999,
			stopIDs: []uint{1, 2},
			userID:  1,
			setupMocks: func(rm *mocks.MockRouteRepository, rsm *mocks.MockRouteStopRepository, sm *mocks.MockStopRepository, rcm *mocks.MockRouteChangeRepository) {
				rm.On("FindByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectError: true,
		},
		{
			name:    "stop not found",
			routeID: 1,
			stopIDs: []uint{1, 999},
			userID:  1,
			setupMocks: func(rm *mocks.MockRouteRepository, rsm *mocks.MockRouteStopRepository, sm *mocks.MockStopRepository, rcm *mocks.MockRouteChangeRepository) {
				rm.On("FindByID", uint(1)).Return(&models.Route{ID: 1}, nil)
				existingStops := []models.RouteStop{}
				rsm.On("FindByRouteID", uint(1)).Return(existingStops, nil)
				sm.On("FindByID", uint(1)).Return(&models.Stop{ID: 1}, nil)
				sm.On("FindByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "stop with ID 999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRouteRepo := mocks.NewMockRouteRepository()
			mockRouteStopRepo := mocks.NewMockRouteStopRepository()
			mockStopRepo := mocks.NewMockStopRepository()
			mockRouteChangeRepo := mocks.NewMockRouteChangeRepository()
			tt.setupMocks(mockRouteRepo, mockRouteStopRepo, mockStopRepo, mockRouteChangeRepo)

			service := NewRouteService(mockRouteRepo, mockRouteStopRepo, mockStopRepo, mockRouteChangeRepo)
			err := service.UpdateRouteStops(tt.routeID, tt.stopIDs, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				// Verify route stops were updated
				mockRouteStopRepo.AssertCalled(t, "DeleteByRouteID", tt.routeID)
				mockRouteStopRepo.AssertNumberOfCalls(t, "Create", len(tt.stopIDs))
			}

			mockRouteRepo.AssertExpectations(t)
			mockRouteStopRepo.AssertExpectations(t)
			mockStopRepo.AssertExpectations(t)
			mockRouteChangeRepo.AssertExpectations(t)
		})
	}
}

func TestRouteService_DeleteRoute(t *testing.T) {
	mockRouteRepo := mocks.NewMockRouteRepository()
	mockRouteStopRepo := mocks.NewMockRouteStopRepository()

	mockRouteStopRepo.On("DeleteByRouteID", uint(1)).Return(nil)
	mockRouteRepo.On("Delete", uint(1)).Return(nil)

	service := NewRouteService(mockRouteRepo, mockRouteStopRepo, mocks.NewMockStopRepository(), mocks.NewMockRouteChangeRepository())
	err := service.DeleteRoute(1)

	assert.NoError(t, err)
	mockRouteRepo.AssertExpectations(t)
	mockRouteStopRepo.AssertExpectations(t)
}

func TestRouteService_ListRoutes(t *testing.T) {
	mockRepo := mocks.NewMockRouteRepository()
	expectedRoutes := []models.Route{
		{ID: 1, RouteNumber: "R9", Name: "Route 9", Status: models.StatusActive},
		{ID: 2, RouteNumber: "R10", Name: "Route 10", Status: models.StatusActive},
	}
	expectedTotal := int64(2)

	filters := map[string]interface{}{
		"status": models.StatusActive,
	}

	mockRepo.On("List", 0, 10, filters).Return(expectedRoutes, expectedTotal, nil)

	service := NewRouteService(mockRepo, mocks.NewMockRouteStopRepository(), mocks.NewMockStopRepository(), mocks.NewMockRouteChangeRepository())
	routes, total, err := service.ListRoutes(0, 10, filters)

	assert.NoError(t, err)
	assert.Equal(t, expectedTotal, total)
	assert.Len(t, routes, 2)
	assert.Equal(t, expectedRoutes[0].RouteNumber, routes[0].RouteNumber)

	mockRepo.AssertExpectations(t)
}

