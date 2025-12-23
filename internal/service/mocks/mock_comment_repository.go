package mocks

import (
	"tj-routes/internal/models"

	"github.com/stretchr/testify/mock"
)

type MockCommentRepository struct {
	mock.Mock
}

func NewMockCommentRepository() *MockCommentRepository {
	return new(MockCommentRepository)
}

func (m *MockCommentRepository) Create(comment *models.Comment) error {
	args := m.Called(comment)
	return args.Error(0)
}

func (m *MockCommentRepository) FindByID(id uint) (*models.Comment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Comment), args.Error(1)
}

func (m *MockCommentRepository) FindByReportID(reportID uint) ([]models.Comment, error) {
	args := m.Called(reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Comment), args.Error(1)
}

func (m *MockCommentRepository) FindThreadedByReportID(reportID uint) ([]models.Comment, error) {
	args := m.Called(reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Comment), args.Error(1)
}

func (m *MockCommentRepository) Update(comment *models.Comment) error {
	args := m.Called(comment)
	return args.Error(0)
}

func (m *MockCommentRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCommentRepository) CountByReportID(reportID uint) (int64, error) {
	args := m.Called(reportID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCommentRepository) CountByUserID(userID uint) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

