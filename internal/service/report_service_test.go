package service

import (
	"testing"

	"tj-routes/internal/models"
	"tj-routes/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestReportService_CreateReport(t *testing.T) {
	tests := []struct {
		name        string
		report      *models.Report
		setupMocks  func(*mocks.MockReportRepository, *mocks.MockRouteRepository, *mocks.MockStopRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful creation",
			report: &models.Report{
				UserID:      1,
				Type:        models.ReportTypeRouteIssue,
				Title:       "Test Report",
				Description: "Test description",
			},
			setupMocks: func(rm *mocks.MockReportRepository, routeRepo *mocks.MockRouteRepository, stopRepo *mocks.MockStopRepository) {
				rm.On("Create", mock.AnythingOfType("*models.Report")).Return(nil).Run(func(args mock.Arguments) {
					report := args.Get(0).(*models.Report)
					assert.Equal(t, models.ReportStatusPending, report.Status)
				})
			},
			expectError: false,
		},
		{
			name: "successful creation with related route",
			report: &models.Report{
				UserID:        1,
				Type:          models.ReportTypeRouteIssue,
				Title:         "Route Issue",
				Description:   "Test description",
				RelatedRouteID: uintPtr(1),
			},
			setupMocks: func(rm *mocks.MockReportRepository, routeRepo *mocks.MockRouteRepository, stopRepo *mocks.MockStopRepository) {
				routeRepo.On("FindByID", uint(1)).Return(&models.Route{ID: 1}, nil)
				rm.On("Create", mock.AnythingOfType("*models.Report")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "successful creation with related stop",
			report: &models.Report{
				UserID:       1,
				Type:         models.ReportTypeStopIssue,
				Title:        "Stop Issue",
				Description:  "Test description",
				RelatedStopID: uintPtr(1),
			},
			setupMocks: func(rm *mocks.MockReportRepository, routeRepo *mocks.MockRouteRepository, stopRepo *mocks.MockStopRepository) {
				stopRepo.On("FindByID", uint(1)).Return(&models.Stop{ID: 1}, nil)
				rm.On("Create", mock.AnythingOfType("*models.Report")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "related route not found",
			report: &models.Report{
				UserID:        1,
				Type:          models.ReportTypeRouteIssue,
				Title:         "Route Issue",
				Description:   "Test description",
				RelatedRouteID: uintPtr(999),
			},
			setupMocks: func(rm *mocks.MockReportRepository, routeRepo *mocks.MockRouteRepository, stopRepo *mocks.MockStopRepository) {
				routeRepo.On("FindByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "related route not found",
		},
		{
			name: "related stop not found",
			report: &models.Report{
				UserID:       1,
				Type:         models.ReportTypeStopIssue,
				Title:        "Stop Issue",
				Description:  "Test description",
				RelatedStopID: uintPtr(999),
			},
			setupMocks: func(rm *mocks.MockReportRepository, routeRepo *mocks.MockRouteRepository, stopRepo *mocks.MockStopRepository) {
				stopRepo.On("FindByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "related stop not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReportRepo := mocks.NewMockReportRepository()
			mockRouteRepo := mocks.NewMockRouteRepository()
			mockStopRepo := mocks.NewMockStopRepository()
			tt.setupMocks(mockReportRepo, mockRouteRepo, mockStopRepo)

			service := NewReportService(mockReportRepo, mockRouteRepo, mockStopRepo)
			err := service.CreateReport(tt.report)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, models.ReportStatusPending, tt.report.Status)
			}

			mockReportRepo.AssertExpectations(t)
			mockRouteRepo.AssertExpectations(t)
			mockStopRepo.AssertExpectations(t)
		})
	}
}

func TestReportService_GetReportByID(t *testing.T) {
	mockRepo := mocks.NewMockReportRepository()
	expectedReport := &models.Report{
		ID:          1,
		UserID:      1,
		Type:        models.ReportTypeRouteIssue,
		Title:       "Test Report",
		Description: "Test description",
		Status:      models.ReportStatusPending,
	}

	mockRepo.On("FindByID", uint(1)).Return(expectedReport, nil)

	service := NewReportService(mockRepo, mocks.NewMockRouteRepository(), mocks.NewMockStopRepository())
	report, err := service.GetReportByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, expectedReport.ID, report.ID)
	assert.Equal(t, expectedReport.Title, report.Title)

	mockRepo.AssertExpectations(t)
}

func TestReportService_UpdateReportStatus(t *testing.T) {
	tests := []struct {
		name           string
		reportID       uint
		currentStatus  models.ReportStatus
		newStatus      models.ReportStatus
		adminNotes     *string
		setupMock      func(*mocks.MockReportRepository)
		expectError    bool
		errorMsg       string
	}{
		{
			name:          "valid transition: pending to reviewed",
			reportID:      1,
			currentStatus: models.ReportStatusPending,
			newStatus:     models.ReportStatusReviewed,
			setupMock: func(m *mocks.MockReportRepository) {
				report := &models.Report{
					ID:     1,
					Status: models.ReportStatusPending,
				}
				m.On("FindByID", uint(1)).Return(report, nil)
				m.On("Update", mock.AnythingOfType("*models.Report")).Return(nil)
			},
			expectError: false,
		},
		{
			name:          "valid transition: pending to resolved",
			reportID:      1,
			currentStatus: models.ReportStatusPending,
			newStatus:     models.ReportStatusResolved,
			setupMock: func(m *mocks.MockReportRepository) {
				report := &models.Report{
					ID:     1,
					Status: models.ReportStatusPending,
				}
				m.On("FindByID", uint(1)).Return(report, nil)
				m.On("Update", mock.AnythingOfType("*models.Report")).Return(nil)
			},
			expectError: false,
		},
		{
			name:          "valid transition: reviewed to resolved",
			reportID:      1,
			currentStatus: models.ReportStatusReviewed,
			newStatus:     models.ReportStatusResolved,
			setupMock: func(m *mocks.MockReportRepository) {
				report := &models.Report{
					ID:     1,
					Status: models.ReportStatusReviewed,
				}
				m.On("FindByID", uint(1)).Return(report, nil)
				m.On("Update", mock.AnythingOfType("*models.Report")).Return(nil)
			},
			expectError: false,
		},
		{
			name:          "invalid transition: resolved cannot change",
			reportID:      1,
			currentStatus: models.ReportStatusResolved,
			newStatus:     models.ReportStatusPending,
			setupMock: func(m *mocks.MockReportRepository) {
				report := &models.Report{
					ID:     1,
					Status: models.ReportStatusResolved,
				}
				m.On("FindByID", uint(1)).Return(report, nil)
			},
			expectError: true,
			errorMsg:    "invalid status transition",
		},
		{
			name:          "invalid transition: reviewed to pending",
			reportID:      1,
			currentStatus: models.ReportStatusReviewed,
			newStatus:     models.ReportStatusPending,
			setupMock: func(m *mocks.MockReportRepository) {
				report := &models.Report{
					ID:     1,
					Status: models.ReportStatusReviewed,
				}
				m.On("FindByID", uint(1)).Return(report, nil)
			},
			expectError: true,
			errorMsg:    "invalid status transition",
		},
		{
			name:          "update with admin notes",
			reportID:      1,
			currentStatus: models.ReportStatusPending,
			newStatus:     models.ReportStatusReviewed,
			adminNotes:    stringPtr("Reviewed by admin"),
			setupMock: func(m *mocks.MockReportRepository) {
				report := &models.Report{
					ID:     1,
					Status: models.ReportStatusPending,
				}
				m.On("FindByID", uint(1)).Return(report, nil)
				m.On("Update", mock.AnythingOfType("*models.Report")).Return(nil).Run(func(args mock.Arguments) {
					report := args.Get(0).(*models.Report)
					assert.NotNil(t, report.AdminNotes)
					assert.Equal(t, "Reviewed by admin", *report.AdminNotes)
				})
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockReportRepository()
			tt.setupMock(mockRepo)

			service := NewReportService(mockRepo, mocks.NewMockRouteRepository(), mocks.NewMockStopRepository())
			err := service.UpdateReportStatus(tt.reportID, tt.newStatus, tt.adminNotes)

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

func TestReportService_DeleteReport(t *testing.T) {
	mockRepo := mocks.NewMockReportRepository()
	mockRepo.On("Delete", uint(1)).Return(nil)

	service := NewReportService(mockRepo, mocks.NewMockRouteRepository(), mocks.NewMockStopRepository())
	err := service.DeleteReport(1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestReportService_ListReports(t *testing.T) {
	mockRepo := mocks.NewMockReportRepository()
	expectedReports := []models.Report{
		{ID: 1, Title: "Report 1", Status: models.ReportStatusPending},
		{ID: 2, Title: "Report 2", Status: models.ReportStatusReviewed},
	}
	expectedTotal := int64(2)

	filters := map[string]interface{}{
		"status": models.ReportStatusPending,
	}

	mockRepo.On("List", 0, 10, filters).Return(expectedReports, expectedTotal, nil)

	service := NewReportService(mockRepo, mocks.NewMockRouteRepository(), mocks.NewMockStopRepository())
	reports, total, err := service.ListReports(0, 10, filters)

	assert.NoError(t, err)
	assert.Equal(t, expectedTotal, total)
	assert.Len(t, reports, 2)
	assert.Equal(t, expectedReports[0].Title, reports[0].Title)

	mockRepo.AssertExpectations(t)
}

func uintPtr(u uint) *uint {
	return &u
}

