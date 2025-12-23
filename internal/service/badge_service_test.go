package service

import (
	"testing"

	"tj-routes/internal/models"
	"tj-routes/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBadgeService_GetAllBadges(t *testing.T) {
	mockBadgeRepo := mocks.NewMockBadgeRepository()
	expectedBadges := []models.Badge{
		{ID: 1, Name: "First Report", CriteriaType: models.BadgeCriteriaReportsAccepted, CriteriaValue: 1},
		{ID: 2, Name: "Helpful Contributor", CriteriaType: models.BadgeCriteriaReportsAccepted, CriteriaValue: 5},
	}

	mockBadgeRepo.On("FindAll").Return(expectedBadges, nil)

	service := NewBadgeService(
		mockBadgeRepo,
		mocks.NewMockUserBadgeRepository(),
		mocks.NewMockUserRepository(nil),
		mocks.NewMockReportRepository(),
		mocks.NewMockCommentRepository(),
		mocks.NewMockReactionRepository(),
	)

	badges, err := service.GetAllBadges()

	assert.NoError(t, err)
	assert.Len(t, badges, 2)
	mockBadgeRepo.AssertExpectations(t)
}

func TestBadgeService_GetUserBadges(t *testing.T) {
	mockUserBadgeRepo := mocks.NewMockUserBadgeRepository()
	expectedUserBadges := []models.UserBadge{
		{ID: 1, UserID: 1, BadgeID: 1},
		{ID: 2, UserID: 1, BadgeID: 2},
	}

	mockUserBadgeRepo.On("FindByUserID", uint(1)).Return(expectedUserBadges, nil)

	service := NewBadgeService(
		mocks.NewMockBadgeRepository(),
		mockUserBadgeRepo,
		mocks.NewMockUserRepository(nil),
		mocks.NewMockReportRepository(),
		mocks.NewMockCommentRepository(),
		mocks.NewMockReactionRepository(),
	)

	badges, err := service.GetUserBadges(1)

	assert.NoError(t, err)
	assert.Len(t, badges, 2)
	mockUserBadgeRepo.AssertExpectations(t)
}

func TestBadgeService_CheckAndAwardBadges(t *testing.T) {
	tests := []struct {
		name        string
		userID      uint
		setupMocks  func(*mocks.MockBadgeRepository, *mocks.MockUserBadgeRepository, *mocks.MockUserRepository, *mocks.MockReportRepository, *mocks.MockCommentRepository, *mocks.MockReactionRepository)
		expectError bool
	}{
		{
			name:   "award badge for reports accepted",
			userID: 1,
			setupMocks: func(bm *mocks.MockBadgeRepository, ubm *mocks.MockUserBadgeRepository, um *mocks.MockUserRepository, rm *mocks.MockReportRepository, cm *mocks.MockCommentRepository, reactm *mocks.MockReactionRepository) {
				badges := []models.Badge{
					{ID: 1, Name: "First Report", CriteriaType: models.BadgeCriteriaReportsAccepted, CriteriaValue: 1},
				}
				bm.On("FindAll").Return(badges, nil)
				ubm.On("FindByUserID", uint(1)).Return([]models.UserBadge{}, nil)
				reports := []models.Report{
					{ID: 1, UserID: 1, Status: models.ReportStatusResolved},
				}
				rm.On("List", 0, 1000, mock.MatchedBy(func(filters map[string]interface{}) bool {
					return filters["user_id"].(uint) == 1 && filters["status"].(models.ReportStatus) == models.ReportStatusResolved
				})).Return(reports, int64(1), nil)
				ubm.On("Create", mock.AnythingOfType("*models.UserBadge")).Return(nil)
			},
			expectError: false,
		},
		{
			name:   "award badge for reputation points",
			userID: 1,
			setupMocks: func(bm *mocks.MockBadgeRepository, ubm *mocks.MockUserBadgeRepository, um *mocks.MockUserRepository, rm *mocks.MockReportRepository, cm *mocks.MockCommentRepository, reactm *mocks.MockReactionRepository) {
				badges := []models.Badge{
					{ID: 1, Name: "Rising Star", CriteriaType: models.BadgeCriteriaReputationPoints, CriteriaValue: 50},
				}
				bm.On("FindAll").Return(badges, nil)
				ubm.On("FindByUserID", uint(1)).Return([]models.UserBadge{}, nil)
				user := &models.User{ID: 1, ReputationPoints: 60}
				um.On("FindByID", uint(1)).Return(user, nil)
				ubm.On("Create", mock.AnythingOfType("*models.UserBadge")).Return(nil)
			},
			expectError: false,
		},
		{
			name:   "skip already awarded badge",
			userID: 1,
			setupMocks: func(bm *mocks.MockBadgeRepository, ubm *mocks.MockUserBadgeRepository, um *mocks.MockUserRepository, rm *mocks.MockReportRepository, cm *mocks.MockCommentRepository, reactm *mocks.MockReactionRepository) {
				badges := []models.Badge{
					{ID: 1, Name: "First Report", CriteriaType: models.BadgeCriteriaReportsAccepted, CriteriaValue: 1},
				}
				bm.On("FindAll").Return(badges, nil)
				existingBadges := []models.UserBadge{
					{ID: 1, UserID: 1, BadgeID: 1},
				}
				ubm.On("FindByUserID", uint(1)).Return(existingBadges, nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBadgeRepo := mocks.NewMockBadgeRepository()
			mockUserBadgeRepo := mocks.NewMockUserBadgeRepository()
			mockUserRepo := mocks.NewMockUserRepository(nil)
			mockReportRepo := mocks.NewMockReportRepository()
			mockCommentRepo := mocks.NewMockCommentRepository()
			mockReactionRepo := mocks.NewMockReactionRepository()

			tt.setupMocks(mockBadgeRepo, mockUserBadgeRepo, mockUserRepo, mockReportRepo, mockCommentRepo, mockReactionRepo)

			service := NewBadgeService(
				mockBadgeRepo,
				mockUserBadgeRepo,
				mockUserRepo,
				mockReportRepo,
				mockCommentRepo,
				mockReactionRepo,
			)

			err := service.CheckAndAwardBadges(tt.userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockBadgeRepo.AssertExpectations(t)
			mockUserBadgeRepo.AssertExpectations(t)
		})
	}
}

