package service

import (
	"testing"

	"tj-routes/internal/models"
	"tj-routes/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestReputationService_CalculateLevel(t *testing.T) {
	service := NewReputationService(mocks.NewMockUserRepository(nil))

	tests := []struct {
		name     string
		points   int
		expected string
	}{
		{"newcomer - 0 points", 0, "newcomer"},
		{"newcomer - 49 points", 49, "newcomer"},
		{"contributor - 50 points", 50, "contributor"},
		{"contributor - 199 points", 199, "contributor"},
		{"trusted - 200 points", 200, "trusted"},
		{"trusted - 499 points", 499, "trusted"},
		{"expert - 500 points", 500, "expert"},
		{"expert - 999 points", 999, "expert"},
		{"legend - 1000 points", 1000, "legend"},
		{"legend - 2000 points", 2000, "legend"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := service.CalculateLevel(tt.points)
			assert.Equal(t, tt.expected, level)
		})
	}
}

func TestReputationService_AddPoints(t *testing.T) {
	tests := []struct {
		name           string
		userID         uint
		initialPoints  int
		initialLevel   string
		pointsToAdd   int
		expectedPoints int
		expectedLevel  string
		setupMock      func(*mocks.MockUserRepository)
		expectError    bool
	}{
		{
			name:          "add points to newcomer",
			userID:        1,
			initialPoints: 10,
			initialLevel:  "newcomer",
			pointsToAdd:   40,
			expectedPoints: 50,
			expectedLevel: "contributor",
			setupMock: func(m *mocks.MockUserRepository) {
				user := &models.User{
					ID:              1,
					ReputationPoints: 10,
					Level:           "newcomer",
				}
				m.On("FindByID", uint(1)).Return(user, nil)
				m.On("Update", mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
					user := args.Get(0).(*models.User)
					assert.Equal(t, 50, user.ReputationPoints)
					assert.Equal(t, "contributor", user.Level)
				})
			},
			expectError: false,
		},
		{
			name:          "add points to reach legend",
			userID:        1,
			initialPoints: 950,
			initialLevel:  "expert",
			pointsToAdd:   50,
			expectedPoints: 1000,
			expectedLevel: "legend",
			setupMock: func(m *mocks.MockUserRepository) {
				user := &models.User{
					ID:              1,
					ReputationPoints: 950,
					Level:           "expert",
				}
				m.On("FindByID", uint(1)).Return(user, nil)
				m.On("Update", mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectError: false,
		},
		{
			name:          "user not found",
			userID:        999,
			initialPoints: 0,
			pointsToAdd:   10,
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("FindByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockUserRepository(nil)
			tt.setupMock(mockRepo)

			service := NewReputationService(mockRepo)
			err := service.AddPoints(tt.userID, tt.pointsToAdd)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestReputationService_UpdateUserLevel(t *testing.T) {
	tests := []struct {
		name          string
		userID        uint
		points        int
		expectedLevel string
		setupMock     func(*mocks.MockUserRepository)
		expectError   bool
	}{
		{
			name:          "update level based on points",
			userID:        1,
			points:        250,
			expectedLevel: "trusted",
			setupMock: func(m *mocks.MockUserRepository) {
				user := &models.User{
					ID:              1,
					ReputationPoints: 250,
					Level:           "contributor",
				}
				m.On("FindByID", uint(1)).Return(user, nil)
				m.On("Update", mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
					user := args.Get(0).(*models.User)
					assert.Equal(t, "trusted", user.Level)
				})
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockUserRepository(nil)
			tt.setupMock(mockRepo)

			service := NewReputationService(mockRepo)
			err := service.UpdateUserLevel(tt.userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestReputationService_GetUserReputation(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(nil)
	user := &models.User{
		ID:              1,
		ReputationPoints: 150,
		Level:           "contributor",
	}
	mockRepo.On("FindByID", uint(1)).Return(user, nil)

	service := NewReputationService(mockRepo)
	points, level, err := service.GetUserReputation(1)

	assert.NoError(t, err)
	assert.Equal(t, 150, points)
	assert.Equal(t, "contributor", level)
	mockRepo.AssertExpectations(t)
}

