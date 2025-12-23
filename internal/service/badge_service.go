package service

import (
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type BadgeService interface {
	CheckAndAwardBadges(userID uint) error
	GetUserBadges(userID uint) ([]models.UserBadge, error)
	GetAllBadges() ([]models.Badge, error)
}

type badgeService struct {
	badgeRepo     repository.BadgeRepository
	userBadgeRepo repository.UserBadgeRepository
	userRepo      repository.UserRepository
	reportRepo    repository.ReportRepository
	commentRepo   repository.CommentRepository
	reactionRepo  repository.ReactionRepository
}

func NewBadgeService(
	badgeRepo repository.BadgeRepository,
	userBadgeRepo repository.UserBadgeRepository,
	userRepo repository.UserRepository,
	reportRepo repository.ReportRepository,
	commentRepo repository.CommentRepository,
	reactionRepo repository.ReactionRepository,
) BadgeService {
	return &badgeService{
		badgeRepo:     badgeRepo,
		userBadgeRepo: userBadgeRepo,
		userRepo:      userRepo,
		reportRepo:    reportRepo,
		commentRepo:   commentRepo,
		reactionRepo:  reactionRepo,
	}
}

func (s *badgeService) CheckAndAwardBadges(userID uint) error {
	// Get all badges
	badges, err := s.badgeRepo.FindAll()
	if err != nil {
		return err
	}

	// Get user's existing badges
	existingBadges, err := s.userBadgeRepo.FindByUserID(userID)
	if err != nil {
		return err
	}

	existingBadgeMap := make(map[uint]bool)
	for _, ub := range existingBadges {
		existingBadgeMap[ub.BadgeID] = true
	}

	// Check each badge criteria
	for _, badge := range badges {
		// Skip if user already has this badge
		if existingBadgeMap[badge.ID] {
			continue
		}

		// Check if user meets criteria
		if s.meetsCriteria(userID, badge) {
			// Award badge
			userBadge := &models.UserBadge{
				UserID:   userID,
				BadgeID:  badge.ID,
			}
			if err := s.userBadgeRepo.Create(userBadge); err != nil {
				// Continue on error - don't fail entire operation
				continue
			}
		}
	}

	return nil
}

func (s *badgeService) meetsCriteria(userID uint, badge models.Badge) bool {
	switch badge.CriteriaType {
	case models.BadgeCriteriaReportsAccepted:
		// Count resolved reports by user
		reports, _, err := s.reportRepo.List(0, 1000, map[string]interface{}{
			"user_id": userID,
			"status":  models.ReportStatusResolved,
		})
		if err != nil {
			return false
		}
		return len(reports) >= badge.CriteriaValue

	case models.BadgeCriteriaCommentsMade:
		count, err := s.commentRepo.CountByUserID(userID)
		if err != nil {
			return false
		}
		return count >= int64(badge.CriteriaValue)

	case models.BadgeCriteriaUpvotesReceived:
		count, err := s.reactionRepo.CountUpvotesByUserContent(userID)
		if err != nil {
			return false
		}
		return count >= int64(badge.CriteriaValue)

	case models.BadgeCriteriaReputationPoints:
		user, err := s.userRepo.FindByID(userID)
		if err != nil {
			return false
		}
		return user.ReputationPoints >= badge.CriteriaValue

	default:
		return false
	}
}

func (s *badgeService) GetUserBadges(userID uint) ([]models.UserBadge, error) {
	return s.userBadgeRepo.FindByUserID(userID)
}

func (s *badgeService) GetAllBadges() ([]models.Badge, error) {
	return s.badgeRepo.FindAll()
}

