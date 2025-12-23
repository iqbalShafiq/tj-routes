package service

import (
	"testing"

	"tj-routes/internal/models"
	"tj-routes/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestReactionService_ToggleReaction(t *testing.T) {
	tests := []struct {
		name           string
		userID         uint
		targetType     models.ReactionTargetType
		targetID       uint
		reactionType   models.ReactionType
		existingReaction *models.Reaction
		setupMocks     func(*mocks.MockReactionRepository, *mocks.MockReportRepository, *mocks.MockCommentRepository, *mocks.MockUserRepository)
		expectError    bool
	}{
		{
			name:         "create new upvote on report",
			userID:       1,
			targetType:   models.ReactionTargetReport,
			targetID:     1,
			reactionType: models.ReactionUpvote,
			setupMocks: func(rm *mocks.MockReactionRepository, reportRepo *mocks.MockReportRepository, commentRepo *mocks.MockCommentRepository, userRepo *mocks.MockUserRepository) {
				rm.On("FindByUserAndTarget", uint(1), models.ReactionTargetReport, uint(1)).Return(nil, gorm.ErrRecordNotFound)
				rm.On("Create", mock.AnythingOfType("*models.Reaction")).Return(nil)
				report := &models.Report{ID: 1, UserID: 2, Upvotes: 0}
				reportRepo.On("FindByID", uint(1)).Return(report, nil)
				reportRepo.On("Update", mock.AnythingOfType("*models.Report")).Return(nil)
				userRepo.On("FindByID", uint(2)).Return(&models.User{ID: 2}, nil)
				userRepo.On("Update", mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectError: false,
		},
		{
			name:         "toggle existing reaction - same type removes",
			userID:       1,
			targetType:   models.ReactionTargetReport,
			targetID:     1,
			reactionType: models.ReactionUpvote,
			existingReaction: &models.Reaction{
				ID:           1,
				UserID:      1,
				TargetType:  models.ReactionTargetReport,
				TargetID:    1,
				ReactionType: models.ReactionUpvote,
			},
			setupMocks: func(rm *mocks.MockReactionRepository, reportRepo *mocks.MockReportRepository, commentRepo *mocks.MockCommentRepository, userRepo *mocks.MockUserRepository) {
				rm.On("FindByUserAndTarget", uint(1), models.ReactionTargetReport, uint(1)).Return(&models.Reaction{
					ID:           1,
					ReactionType: models.ReactionUpvote,
				}, nil)
				rm.On("DeleteByUserAndTarget", uint(1), models.ReactionTargetReport, uint(1)).Return(nil)
				report := &models.Report{ID: 1, UserID: 2, Upvotes: 1}
				reportRepo.On("FindByID", uint(1)).Return(report, nil)
				reportRepo.On("Update", mock.AnythingOfType("*models.Report")).Return(nil)
				userRepo.On("FindByID", uint(2)).Return(&models.User{ID: 2}, nil)
				userRepo.On("Update", mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectError: false,
		},
		{
			name:         "change reaction type - upvote to downvote",
			userID:       1,
			targetType:   models.ReactionTargetReport,
			targetID:     1,
			reactionType: models.ReactionDownvote,
			existingReaction: &models.Reaction{
				ID:           1,
				ReactionType: models.ReactionUpvote,
			},
			setupMocks: func(rm *mocks.MockReactionRepository, reportRepo *mocks.MockReportRepository, commentRepo *mocks.MockCommentRepository, userRepo *mocks.MockUserRepository) {
				rm.On("FindByUserAndTarget", uint(1), models.ReactionTargetReport, uint(1)).Return(&models.Reaction{
					ID:           1,
					ReactionType: models.ReactionUpvote,
				}, nil)
				rm.On("Update", mock.AnythingOfType("*models.Reaction")).Return(nil)
				report := &models.Report{ID: 1, UserID: 2, Upvotes: 1, Downvotes: 0}
				reportRepo.On("FindByID", uint(1)).Return(report, nil).Times(2)
				reportRepo.On("Update", mock.AnythingOfType("*models.Report")).Return(nil).Times(2)
				userRepo.On("FindByID", uint(2)).Return(&models.User{ID: 2}, nil).Times(2)
				userRepo.On("Update", mock.AnythingOfType("*models.User")).Return(nil).Times(2)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReactionRepo := mocks.NewMockReactionRepository()
			mockReportRepo := mocks.NewMockReportRepository()
			mockCommentRepo := mocks.NewMockCommentRepository()
			mockUserRepo := mocks.NewMockUserRepository(nil)
			mockReputationService := NewReputationService(mockUserRepo)

			tt.setupMocks(mockReactionRepo, mockReportRepo, mockCommentRepo, mockUserRepo)

			service := NewReactionService(mockReactionRepo, mockReportRepo, mockCommentRepo, mockReputationService)
			err := service.ToggleReaction(tt.userID, tt.targetType, tt.targetID, tt.reactionType)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockReactionRepo.AssertExpectations(t)
		})
	}
}

func TestReactionService_RemoveReaction(t *testing.T) {
	mockReactionRepo := mocks.NewMockReactionRepository()
	mockReportRepo := mocks.NewMockReportRepository()
	mockCommentRepo := mocks.NewMockCommentRepository()
	mockUserRepo := mocks.NewMockUserRepository(nil)
	mockReputationService := NewReputationService(mockUserRepo)

	existingReaction := &models.Reaction{
		ID:           1,
		UserID:       1,
		TargetType:   models.ReactionTargetReport,
		TargetID:     1,
		ReactionType: models.ReactionUpvote,
	}

	mockReactionRepo.On("FindByUserAndTarget", uint(1), models.ReactionTargetReport, uint(1)).Return(existingReaction, nil)
	mockReactionRepo.On("DeleteByUserAndTarget", uint(1), models.ReactionTargetReport, uint(1)).Return(nil)
	report := &models.Report{ID: 1, UserID: 2, Upvotes: 1}
	mockReportRepo.On("FindByID", uint(1)).Return(report, nil)
	mockReportRepo.On("Update", mock.AnythingOfType("*models.Report")).Return(nil)
	mockUserRepo.On("FindByID", uint(2)).Return(&models.User{ID: 2}, nil)
	mockUserRepo.On("Update", mock.AnythingOfType("*models.User")).Return(nil)

	service := NewReactionService(mockReactionRepo, mockReportRepo, mockCommentRepo, mockReputationService)
	err := service.RemoveReaction(1, models.ReactionTargetReport, 1)

	assert.NoError(t, err)
	mockReactionRepo.AssertExpectations(t)
}

func TestReactionService_GetUserReaction(t *testing.T) {
	mockReactionRepo := mocks.NewMockReactionRepository()
	mockReportRepo := mocks.NewMockReportRepository()
	mockCommentRepo := mocks.NewMockCommentRepository()
	mockUserRepo := mocks.NewMockUserRepository(nil)
	mockReputationService := NewReputationService(mockUserRepo)

	expectedReaction := &models.Reaction{
		ID:           1,
		UserID:       1,
		TargetType:   models.ReactionTargetReport,
		TargetID:     1,
		ReactionType: models.ReactionUpvote,
	}

	mockReactionRepo.On("FindByUserAndTarget", uint(1), models.ReactionTargetReport, uint(1)).Return(expectedReaction, nil)

	service := NewReactionService(mockReactionRepo, mockReportRepo, mockCommentRepo, mockReputationService)
	reaction, err := service.GetUserReaction(1, models.ReactionTargetReport, 1)

	assert.NoError(t, err)
	assert.NotNil(t, reaction)
	assert.Equal(t, models.ReactionUpvote, reaction.ReactionType)
	mockReactionRepo.AssertExpectations(t)
}

