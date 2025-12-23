package mocks

import (
	"tj-routes/internal/models"

	"github.com/stretchr/testify/mock"
)

type MockReactionRepository struct {
	mock.Mock
}

func NewMockReactionRepository() *MockReactionRepository {
	return new(MockReactionRepository)
}

func (m *MockReactionRepository) Create(reaction *models.Reaction) error {
	args := m.Called(reaction)
	return args.Error(0)
}

func (m *MockReactionRepository) Update(reaction *models.Reaction) error {
	args := m.Called(reaction)
	return args.Error(0)
}

func (m *MockReactionRepository) FindByUserAndTarget(userID uint, targetType models.ReactionTargetType, targetID uint) (*models.Reaction, error) {
	args := m.Called(userID, targetType, targetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reaction), args.Error(1)
}

func (m *MockReactionRepository) FindByTarget(targetType models.ReactionTargetType, targetID uint) ([]models.Reaction, error) {
	args := m.Called(targetType, targetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Reaction), args.Error(1)
}

func (m *MockReactionRepository) CountByTargetAndType(targetType models.ReactionTargetType, targetID uint, reactionType models.ReactionType) (int64, error) {
	args := m.Called(targetType, targetID, reactionType)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockReactionRepository) CountUpvotesByUserContent(userID uint) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockReactionRepository) Delete(reactionID uint) error {
	args := m.Called(reactionID)
	return args.Error(0)
}

func (m *MockReactionRepository) DeleteByUserAndTarget(userID uint, targetType models.ReactionTargetType, targetID uint) error {
	args := m.Called(userID, targetType, targetID)
	return args.Error(0)
}

