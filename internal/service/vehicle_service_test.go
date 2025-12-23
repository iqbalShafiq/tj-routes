package service

import (
	"testing"

	"tj-routes/internal/models"
	"tj-routes/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestVehicleService_CreateVehicle(t *testing.T) {
	tests := []struct {
		name        string
		vehicle     *models.Vehicle
		setupMocks  func(*mocks.MockVehicleRepository, *mocks.MockRouteRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful creation",
			vehicle: &models.Vehicle{
				VehiclePlate: "B1234XYZ",
				RouteID:      1,
				VehicleType:  "Bus",
				Capacity:     50,
				Status:       models.StatusActive,
			},
			setupMocks: func(vm *mocks.MockVehicleRepository, rm *mocks.MockRouteRepository) {
				rm.On("FindByID", uint(1)).Return(&models.Route{ID: 1}, nil)
				vm.On("Create", mock.AnythingOfType("*models.Vehicle")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "route not found",
			vehicle: &models.Vehicle{
				VehiclePlate: "B1234XYZ",
				RouteID:      999,
				VehicleType:  "Bus",
				Capacity:     50,
			},
			setupMocks: func(vm *mocks.MockVehicleRepository, rm *mocks.MockRouteRepository) {
				rm.On("FindByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockVehicleRepo := mocks.NewMockVehicleRepository()
			mockRouteRepo := mocks.NewMockRouteRepository()
			tt.setupMocks(mockVehicleRepo, mockRouteRepo)

			service := NewVehicleService(mockVehicleRepo, mockRouteRepo, getTestCache(), getTestConfig())
			err := service.CreateVehicle(tt.vehicle)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockVehicleRepo.AssertExpectations(t)
			mockRouteRepo.AssertExpectations(t)
		})
	}
}

func TestVehicleService_GetVehicleByID(t *testing.T) {
	mockRepo := mocks.NewMockVehicleRepository()
	expectedVehicle := &models.Vehicle{
		ID:           1,
		VehiclePlate: "B1234XYZ",
		RouteID:      1,
		VehicleType:  "Bus",
		Capacity:     50,
		Status:       models.StatusActive,
	}

	mockRepo.On("FindByID", uint(1)).Return(expectedVehicle, nil)

	service := NewVehicleService(mockRepo, mocks.NewMockRouteRepository(), getTestCache(), getTestConfig())
	vehicle, err := service.GetVehicleByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, vehicle)
	assert.Equal(t, expectedVehicle.ID, vehicle.ID)
	assert.Equal(t, expectedVehicle.VehiclePlate, vehicle.VehiclePlate)

	mockRepo.AssertExpectations(t)
}

func TestVehicleService_UpdateVehicle(t *testing.T) {
	tests := []struct {
		name        string
		vehicle     *models.Vehicle
		setupMocks  func(*mocks.MockVehicleRepository, *mocks.MockRouteRepository)
		expectError bool
	}{
		{
			name: "successful update",
			vehicle: &models.Vehicle{
				ID:           1,
				VehiclePlate: "B1234XYZ",
				RouteID:      1,
				Capacity:     60,
			},
			setupMocks: func(vm *mocks.MockVehicleRepository, rm *mocks.MockRouteRepository) {
				rm.On("FindByID", uint(1)).Return(&models.Route{ID: 1}, nil)
				vm.On("Update", mock.AnythingOfType("*models.Vehicle")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "route not found",
			vehicle: &models.Vehicle{
				ID:           1,
				VehiclePlate: "B1234XYZ",
				RouteID:      999,
			},
			setupMocks: func(vm *mocks.MockVehicleRepository, rm *mocks.MockRouteRepository) {
				rm.On("FindByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockVehicleRepo := mocks.NewMockVehicleRepository()
			mockRouteRepo := mocks.NewMockRouteRepository()
			tt.setupMocks(mockVehicleRepo, mockRouteRepo)

			service := NewVehicleService(mockVehicleRepo, mockRouteRepo, getTestCache(), getTestConfig())
			err := service.UpdateVehicle(tt.vehicle)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockVehicleRepo.AssertExpectations(t)
			mockRouteRepo.AssertExpectations(t)
		})
	}
}

func TestVehicleService_DeleteVehicle(t *testing.T) {
	mockRepo := mocks.NewMockVehicleRepository()
	mockRepo.On("Delete", uint(1)).Return(nil)

	service := NewVehicleService(mockRepo, mocks.NewMockRouteRepository(), getTestCache(), getTestConfig())
	err := service.DeleteVehicle(1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestVehicleService_ListVehicles(t *testing.T) {
	mockRepo := mocks.NewMockVehicleRepository()
	expectedVehicles := []models.Vehicle{
		{ID: 1, VehiclePlate: "B1234XYZ", RouteID: 1},
		{ID: 2, VehiclePlate: "B5678ABC", RouteID: 1},
	}
	expectedTotal := int64(2)

	filters := map[string]interface{}{
		"route_id": uint(1),
	}

	mockRepo.On("List", 0, 10, filters).Return(expectedVehicles, expectedTotal, nil)

	service := NewVehicleService(mockRepo, mocks.NewMockRouteRepository(), getTestCache(), getTestConfig())
	vehicles, total, err := service.ListVehicles(0, 10, filters)

	assert.NoError(t, err)
	assert.Equal(t, expectedTotal, total)
	assert.Len(t, vehicles, 2)
	assert.Equal(t, expectedVehicles[0].VehiclePlate, vehicles[0].VehiclePlate)

	mockRepo.AssertExpectations(t)
}

