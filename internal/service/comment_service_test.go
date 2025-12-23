package service

import (
	"testing"

	"tj-routes/internal/models"
	"tj-routes/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestCommentService_CreateComment(t *testing.T) {
	tests := []struct {
		name        string
		comment     *models.Comment
		setupMocks  func(*mocks.MockCommentRepository, *mocks.MockReportRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful creation - top level comment",
			comment: &models.Comment{
				ReportID: 1,
				UserID:   1,
				Content:  "Test comment",
				ParentID: nil,
			},
			setupMocks: func(cm *mocks.MockCommentRepository, rm *mocks.MockReportRepository) {
				report := &models.Report{ID: 1, CommentCount: 0}
				rm.On("FindByID", uint(1)).Return(report, nil)
				cm.On("Create", mock.AnythingOfType("*models.Comment")).Return(nil)
				cm.On("CountByReportID", uint(1)).Return(int64(1), nil)
				rm.On("Update", mock.AnythingOfType("*models.Report")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "successful creation - reply comment",
			comment: &models.Comment{
				ReportID: 1,
				UserID:   1,
				Content:  "Reply comment",
				ParentID: uintPtr(10),
			},
			setupMocks: func(cm *mocks.MockCommentRepository, rm *mocks.MockReportRepository) {
				report := &models.Report{ID: 1, CommentCount: 0}
				parentComment := &models.Comment{ID: 10, ReportID: 1}
				rm.On("FindByID", uint(1)).Return(report, nil)
				cm.On("FindByID", uint(10)).Return(parentComment, nil)
				cm.On("Create", mock.AnythingOfType("*models.Comment")).Return(nil)
				cm.On("CountByReportID", uint(1)).Return(int64(2), nil)
				rm.On("Update", mock.AnythingOfType("*models.Report")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "report not found",
			comment: &models.Comment{
				ReportID: 999,
				UserID:   1,
				Content:  "Test comment",
			},
			setupMocks: func(cm *mocks.MockCommentRepository, rm *mocks.MockReportRepository) {
				rm.On("FindByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "report not found",
		},
		{
			name: "parent comment not found",
			comment: &models.Comment{
				ReportID: 1,
				UserID:   1,
				Content:  "Reply comment",
				ParentID: uintPtr(999),
			},
			setupMocks: func(cm *mocks.MockCommentRepository, rm *mocks.MockReportRepository) {
				report := &models.Report{ID: 1}
				rm.On("FindByID", uint(1)).Return(report, nil)
				cm.On("FindByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "parent comment not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommentRepo := mocks.NewMockCommentRepository()
			mockReportRepo := mocks.NewMockReportRepository()
			tt.setupMocks(mockCommentRepo, mockReportRepo)

			service := NewCommentService(mockCommentRepo, mockReportRepo)
			err := service.CreateComment(tt.comment)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockCommentRepo.AssertExpectations(t)
			mockReportRepo.AssertExpectations(t)
		})
	}
}

func TestCommentService_GetCommentsByReportID(t *testing.T) {
	mockRepo := mocks.NewMockCommentRepository()
	expectedComments := []models.Comment{
		{ID: 1, ReportID: 1, Content: "Comment 1"},
		{ID: 2, ReportID: 1, Content: "Comment 2"},
	}
	mockRepo.On("FindThreadedByReportID", uint(1)).Return(expectedComments, nil)

	service := NewCommentService(mockRepo, mocks.NewMockReportRepository())
	comments, err := service.GetCommentsByReportID(1)

	assert.NoError(t, err)
	assert.Len(t, comments, 2)
	mockRepo.AssertExpectations(t)
}

func TestCommentService_UpdateComment(t *testing.T) {
	tests := []struct {
		name        string
		commentID   uint
		userID      uint
		content     string
		setupMock   func(*mocks.MockCommentRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:      "successful update",
			commentID: 1,
			userID:    1,
			content:   "Updated content",
			setupMock: func(m *mocks.MockCommentRepository) {
				comment := &models.Comment{ID: 1, UserID: 1, Content: "Old content"}
				m.On("FindByID", uint(1)).Return(comment, nil)
				m.On("Update", mock.AnythingOfType("*models.Comment")).Return(nil)
			},
			expectError: false,
		},
		{
			name:      "unauthorized - different user",
			commentID: 1,
			userID:    2,
			content:   "Updated content",
			setupMock: func(m *mocks.MockCommentRepository) {
				comment := &models.Comment{ID: 1, UserID: 1, Content: "Old content"}
				m.On("FindByID", uint(1)).Return(comment, nil)
			},
			expectError: true,
			errorMsg:    "unauthorized",
		},
		{
			name:      "comment not found",
			commentID: 999,
			userID:    1,
			content:   "Updated content",
			setupMock: func(m *mocks.MockCommentRepository) {
				m.On("FindByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorMsg:    "comment not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockCommentRepository()
			tt.setupMock(mockRepo)

			service := NewCommentService(mockRepo, mocks.NewMockReportRepository())
			err := service.UpdateComment(tt.commentID, tt.userID, tt.content)

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

func TestCommentService_DeleteComment(t *testing.T) {
	tests := []struct {
		name        string
		commentID   uint
		userID      uint
		setupMock   func(*mocks.MockCommentRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:      "successful delete",
			commentID: 1,
			userID:    1,
			setupMock: func(m *mocks.MockCommentRepository) {
				comment := &models.Comment{ID: 1, UserID: 1}
				m.On("FindByID", uint(1)).Return(comment, nil)
				m.On("Delete", uint(1)).Return(nil)
			},
			expectError: false,
		},
		{
			name:      "unauthorized - different user",
			commentID: 1,
			userID:    2,
			setupMock: func(m *mocks.MockCommentRepository) {
				comment := &models.Comment{ID: 1, UserID: 1}
				m.On("FindByID", uint(1)).Return(comment, nil)
			},
			expectError: true,
			errorMsg:    "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockCommentRepository()
			tt.setupMock(mockRepo)

			service := NewCommentService(mockRepo, mocks.NewMockReportRepository())
			err := service.DeleteComment(tt.commentID, tt.userID)

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
