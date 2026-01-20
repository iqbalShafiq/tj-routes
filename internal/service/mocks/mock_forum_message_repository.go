package mocks

import (
	"time"

	"tj-routes/internal/models"

	"github.com/stretchr/testify/mock"
)

type MockForumMessageRepository struct {
	mock.Mock
}

func NewMockForumMessageRepository() *MockForumMessageRepository {
	return new(MockForumMessageRepository)
}

func (m *MockForumMessageRepository) Create(message *models.ForumMessage) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockForumMessageRepository) FindByID(id uint) (*models.ForumMessage, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ForumMessage), args.Error(1)
}

func (m *MockForumMessageRepository) ListByForumID(forumID uint, offset, limit int) ([]models.ForumMessage, int64, error) {
	args := m.Called(forumID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.ForumMessage), args.Get(1).(int64), args.Error(2)
}

func (m *MockForumMessageRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockForumMessageRepository) DeleteOlderThan(cutoff time.Time) error {
	args := m.Called(cutoff)
	return args.Error(0)
}
