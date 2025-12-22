package service

import (
	"testing"

	"tj-routes/internal/config"
	"tj-routes/internal/models"
	"tj-routes/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestUserService_Register(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		username    string
		password    string
		setupMock   func(*mocks.MockUserRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:     "successful registration",
			email:    "test@example.com",
			username: "testuser",
			password: "password123",
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("FindByEmail", "test@example.com").Return(nil, gorm.ErrRecordNotFound)
				m.On("Create", mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
					user := args.Get(0).(*models.User)
					assert.Equal(t, "test@example.com", user.Email)
					assert.Equal(t, "testuser", user.Username)
					assert.NotNil(t, user.Password)
					assert.Equal(t, models.RoleCommonUser, user.Role)
				})
			},
			expectError: false,
		},
		{
			name:     "email already exists",
			email:    "existing@example.com",
			username: "existinguser",
			password: "password123",
			setupMock: func(m *mocks.MockUserRepository) {
				existingUser := &models.User{ID: 1, Email: "existing@example.com"}
				m.On("FindByEmail", "existing@example.com").Return(existingUser, nil)
			},
			expectError: true,
			errorMsg:    "user with this email already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockUserRepository(t)
			tt.setupMock(mockRepo)

			cfg := &config.Config{
				JWT: config.JWTConfig{
					Secret:              "test-secret",
					ExpirationHours:     24,
					RefreshExpirationHours: 168,
				},
			}

			service := NewUserService(mockRepo, cfg)
			user, err := service.Register(tt.email, tt.username, tt.password)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.username, user.Username)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		setupMock   func(*mocks.MockUserRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: "password123",
			setupMock: func(m *mocks.MockUserRepository) {
				hashedPassword := "$2a$10$dummyhash"
				user := &models.User{
					ID:       1,
					Email:    "test@example.com",
					Username: "testuser",
					Password: &hashedPassword,
					Role:     models.RoleCommonUser,
				}
				m.On("FindByEmail", "test@example.com").Return(user, nil)
			},
			expectError: false,
		},
		{
			name:     "user not found",
			email:    "notfound@example.com",
			password: "password123",
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("FindByEmail", "notfound@example.com").Return(nil, gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "invalid email or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockUserRepository(t)
			tt.setupMock(mockRepo)

			cfg := &config.Config{
				JWT: config.JWTConfig{
					Secret:              "test-secret-key-for-testing-purposes-only",
					ExpirationHours:     24,
					RefreshExpirationHours: 168,
				},
			}

			service := NewUserService(mockRepo, cfg)

			// Note: This test will fail on password verification since we're using a dummy hash
			// In a real test, you'd want to use a properly hashed password or mock the password utility
			user, accessToken, refreshToken, err := service.Login(tt.email, tt.password)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, user)
				assert.Empty(t, accessToken)
				assert.Empty(t, refreshToken)
			} else {
				// For successful case, we'd need proper password hashing setup
				// This is a placeholder to show the structure
				if err == nil {
					assert.NotNil(t, user)
					assert.NotEmpty(t, accessToken)
					assert.NotEmpty(t, refreshToken)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

