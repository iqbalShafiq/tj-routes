package mocks

import (
	"tj-routes/internal/models"

	"github.com/stretchr/testify/mock"
)

type MockBadgeRepository struct {
	mock.Mock
}

func NewMockBadgeRepository() *MockBadgeRepository {
	return new(MockBadgeRepository)
}

func (m *MockBadgeRepository) Create(badge *models.Badge) error {
	args := m.Called(badge)
	return args.Error(0)
}

func (m *MockBadgeRepository) FindByID(id uint) (*models.Badge, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Badge), args.Error(1)
}

func (m *MockBadgeRepository) FindAll() ([]models.Badge, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Badge), args.Error(1)
}

func (m *MockBadgeRepository) FindByCriteriaType(criteriaType models.BadgeCriteriaType) ([]models.Badge, error) {
	args := m.Called(criteriaType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Badge), args.Error(1)
}

func (m *MockBadgeRepository) Update(badge *models.Badge) error {
	args := m.Called(badge)
	return args.Error(0)
}

func (m *MockBadgeRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockUserBadgeRepository struct {
	mock.Mock
}

func NewMockUserBadgeRepository() *MockUserBadgeRepository {
	return new(MockUserBadgeRepository)
}

func (m *MockUserBadgeRepository) Create(userBadge *models.UserBadge) error {
	args := m.Called(userBadge)
	return args.Error(0)
}

func (m *MockUserBadgeRepository) FindByUserID(userID uint) ([]models.UserBadge, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserBadge), args.Error(1)
}

func (m *MockUserBadgeRepository) FindByUserAndBadge(userID, badgeID uint) (*models.UserBadge, error) {
	args := m.Called(userID, badgeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserBadge), args.Error(1)
}

func (m *MockUserBadgeRepository) FindByBadgeID(badgeID uint) ([]models.UserBadge, error) {
	args := m.Called(badgeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserBadge), args.Error(1)
}

func (m *MockUserBadgeRepository) Delete(userID, badgeID uint) error {
	args := m.Called(userID, badgeID)
	return args.Error(0)
}

