package mocks

import (
	"tj-routes/internal/models"

	"github.com/stretchr/testify/mock"
)

type MockForumMemberRepository struct {
	mock.Mock
}

func NewMockForumMemberRepository() *MockForumMemberRepository {
	return new(MockForumMemberRepository)
}

func (m *MockForumMemberRepository) Create(member *models.ForumMember) error {
	args := m.Called(member)
	return args.Error(0)
}

func (m *MockForumMemberRepository) FindByForumAndUser(forumID uint, userID uint) (*models.ForumMember, error) {
	args := m.Called(forumID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ForumMember), args.Error(1)
}

func (m *MockForumMemberRepository) Delete(forumID uint, userID uint) error {
	args := m.Called(forumID, userID)
	return args.Error(0)
}

func (m *MockForumMemberRepository) ListMembers(forumID uint, offset, limit int) ([]models.ForumMember, int64, error) {
	args := m.Called(forumID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.ForumMember), args.Get(1).(int64), args.Error(2)
}

func (m *MockForumMemberRepository) ListForumsByUser(userID uint, offset, limit int) ([]models.Forum, int64, error) {
	args := m.Called(userID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Forum), args.Get(1).(int64), args.Error(2)
}

func (m *MockForumMemberRepository) CountByForumID(forumID uint) (int64, error) {
	args := m.Called(forumID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockForumMemberRepository) IsMember(forumID uint, userID uint) (bool, error) {
	args := m.Called(forumID, userID)
	return args.Bool(0), args.Error(1)
}
