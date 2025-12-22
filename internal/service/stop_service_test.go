package service

import (
	"testing"

	"tj-routes/internal/models"
	"tj-routes/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestStopService_CreateStop(t *testing.T) {
	tests := []struct {
		name        string
		stop        *models.Stop
		setupMock   func(*mocks.MockStopRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful creation",
			stop: &models.Stop{
				Name:      "Test Stop",
				Type:      models.StopTypeStop,
				Latitude:  -6.2088,
				Longitude: 106.8456,
				Status:    models.StatusActive,
			},
			setupMock: func(m *mocks.MockStopRepository) {
				m.On("Create", mock.AnythingOfType("*models.Stop")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "successful creation with valid JSON facilities",
			stop: &models.Stop{
				Name:      "Test Stop",
				Type:      models.StopTypeStop,
				Latitude:  -6.2088,
				Longitude: 106.8456,
				Facilities: stringPtr(`{"wifi": true, "parking": true}`),
				Status:    models.StatusActive,
			},
			setupMock: func(m *mocks.MockStopRepository) {
				m.On("Create", mock.AnythingOfType("*models.Stop")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "invalid JSON facilities",
			stop: &models.Stop{
				Name:       "Test Stop",
				Type:       models.StopTypeStop,
				Latitude:   -6.2088,
				Longitude:  106.8456,
				Facilities: stringPtr(`{"invalid": json}`),
				Status:     models.StatusActive,
			},
			setupMock:   func(m *mocks.MockStopRepository) {},
			expectError: true,
			errorMsg:    "invalid facilities JSON",
		},
		{
			name: "database error",
			stop: &models.Stop{
				Name:      "Test Stop",
				Type:      models.StopTypeStop,
				Latitude:  -6.2088,
				Longitude: 106.8456,
				Status:    models.StatusActive,
			},
			setupMock: func(m *mocks.MockStopRepository) {
				m.On("Create", mock.AnythingOfType("*models.Stop")).Return(gorm.ErrInvalidDB)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockStopRepository()
			tt.setupMock(mockRepo)

			service := NewStopService(mockRepo)
			err := service.CreateStop(tt.stop)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestStopService_GetStopByID(t *testing.T) {
	mockRepo := mocks.NewMockStopRepository()
	expectedStop := &models.Stop{
		ID:        1,
		Name:      "Test Stop",
		Type:      models.StopTypeStop,
		Latitude:  -6.2088,
		Longitude: 106.8456,
		Status:    models.StatusActive,
	}

	mockRepo.On("FindByID", uint(1)).Return(expectedStop, nil)

	service := NewStopService(mockRepo)
	stop, err := service.GetStopByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, stop)
	assert.Equal(t, expectedStop.ID, stop.ID)
	assert.Equal(t, expectedStop.Name, stop.Name)

	mockRepo.AssertExpectations(t)
}

func TestStopService_UpdateStop(t *testing.T) {
	tests := []struct {
		name        string
		stop        *models.Stop
		setupMock   func(*mocks.MockStopRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful update",
			stop: &models.Stop{
				ID:        1,
				Name:      "Updated Stop",
				Type:      models.StopTypeStop,
				Latitude:  -6.2088,
				Longitude: 106.8456,
				Status:    models.StatusActive,
			},
			setupMock: func(m *mocks.MockStopRepository) {
				m.On("Update", mock.AnythingOfType("*models.Stop")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "invalid JSON facilities",
			stop: &models.Stop{
				ID:         1,
				Name:       "Test Stop",
				Type:       models.StopTypeStop,
				Latitude:   -6.2088,
				Longitude:  106.8456,
				Facilities: stringPtr(`invalid json`),
				Status:     models.StatusActive,
			},
			setupMock:   func(m *mocks.MockStopRepository) {},
			expectError: true,
			errorMsg:    "invalid facilities JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockStopRepository()
			tt.setupMock(mockRepo)

			service := NewStopService(mockRepo)
			err := service.UpdateStop(tt.stop)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestStopService_DeleteStop(t *testing.T) {
	mockRepo := mocks.NewMockStopRepository()
	mockRepo.On("Delete", uint(1)).Return(nil)

	service := NewStopService(mockRepo)
	err := service.DeleteStop(1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestStopService_ListStops(t *testing.T) {
	mockRepo := mocks.NewMockStopRepository()
	expectedStops := []models.Stop{
		{ID: 1, Name: "Stop 1", Status: models.StatusActive},
		{ID: 2, Name: "Stop 2", Status: models.StatusActive},
	}
	expectedTotal := int64(2)

	filters := map[string]interface{}{
		"status": models.StatusActive,
	}

	mockRepo.On("List", 0, 10, filters).Return(expectedStops, expectedTotal, nil)

	service := NewStopService(mockRepo)
	stops, total, err := service.ListStops(0, 10, filters)

	assert.NoError(t, err)
	assert.Equal(t, expectedTotal, total)
	assert.Len(t, stops, 2)
	assert.Equal(t, expectedStops[0].Name, stops[0].Name)

	mockRepo.AssertExpectations(t)
}

