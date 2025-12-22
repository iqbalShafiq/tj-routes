package mocks

import (
	"tj-routes/internal/models"

	"github.com/stretchr/testify/mock"
)

type MockReportRepository struct {
	mock.Mock
}

func NewMockReportRepository() *MockReportRepository {
	return new(MockReportRepository)
}

func (m *MockReportRepository) Create(report *models.Report) error {
	args := m.Called(report)
	return args.Error(0)
}

func (m *MockReportRepository) FindByID(id uint) (*models.Report, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Report), args.Error(1)
}

func (m *MockReportRepository) Update(report *models.Report) error {
	args := m.Called(report)
	return args.Error(0)
}

func (m *MockReportRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockReportRepository) List(offset, limit int, filters map[string]interface{}) ([]models.Report, int64, error) {
	args := m.Called(offset, limit, filters)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Report), args.Get(1).(int64), args.Error(2)
}

