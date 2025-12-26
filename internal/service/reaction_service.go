package service

import (
	"errors"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type ReactionService interface {
	ToggleReaction(userID uint, targetType models.ReactionTargetType, targetID uint, reactionType models.ReactionType) error
	RemoveReaction(userID uint, targetType models.ReactionTargetType, targetID uint) error
	GetUserReaction(userID uint, targetType models.ReactionTargetType, targetID uint) (*models.Reaction, error)
}

type reactionService struct {
	reactionRepo  repository.ReactionRepository
	reportRepo    repository.ReportRepository
	commentRepo   repository.CommentRepository
	forumPostRepo repository.ForumPostRepository
	reputationService ReputationService
}

func NewReactionService(
	reactionRepo repository.ReactionRepository,
	reportRepo repository.ReportRepository,
	commentRepo repository.CommentRepository,
	reputationService ReputationService,
) ReactionService {
	return &reactionService{
		reactionRepo:      reactionRepo,
		reportRepo:        reportRepo,
		commentRepo:       commentRepo,
		reputationService: reputationService,
	}
}

func NewReactionServiceWithForumPost(
	reactionRepo repository.ReactionRepository,
	reportRepo repository.ReportRepository,
	commentRepo repository.CommentRepository,
	forumPostRepo repository.ForumPostRepository,
	reputationService ReputationService,
) ReactionService {
	return &reactionService{
		reactionRepo:      reactionRepo,
		reportRepo:        reportRepo,
		commentRepo:       commentRepo,
		forumPostRepo:     forumPostRepo,
		reputationService: reputationService,
	}
}

func (s *reactionService) ToggleReaction(userID uint, targetType models.ReactionTargetType, targetID uint, reactionType models.ReactionType) error {
	// Check if user already has a reaction
	existing, err := s.reactionRepo.FindByUserAndTarget(userID, targetType, targetID)
	if err == nil && existing != nil {
		// User already reacted - toggle or change reaction
		if existing.ReactionType == reactionType {
			// Same reaction - remove it
			return s.RemoveReaction(userID, targetType, targetID)
		} else {
			// Different reaction - update it
			oldType := existing.ReactionType
			existing.ReactionType = reactionType
			if err := s.reactionRepo.Update(existing); err != nil {
				return err
			}
			// Update counts - remove old, add new
			s.updateTargetCounts(targetType, targetID, oldType, -1)
			s.updateTargetCounts(targetType, targetID, reactionType, 1)
			// Update reputation for owner
			s.updateOwnerReputation(targetType, targetID, oldType, reactionType)
			return nil
		}
	}

	// Create new reaction
	reaction := &models.Reaction{
		UserID:       userID,
		TargetType:   targetType,
		TargetID:     targetID,
		ReactionType: reactionType,
	}

	if err := s.reactionRepo.Create(reaction); err != nil {
		return err
	}

	// Update target counts
	s.updateTargetCounts(targetType, targetID, reactionType, 1)

	// Award points to content owner
	s.awardReactionPoints(targetType, targetID, reactionType)

	return nil
}

func (s *reactionService) RemoveReaction(userID uint, targetType models.ReactionTargetType, targetID uint) error {
	existing, err := s.reactionRepo.FindByUserAndTarget(userID, targetType, targetID)
	if err != nil {
		return errors.New("reaction not found")
	}

	reactionType := existing.ReactionType
	if err := s.reactionRepo.DeleteByUserAndTarget(userID, targetType, targetID); err != nil {
		return err
	}

	// Update target counts
	s.updateTargetCounts(targetType, targetID, reactionType, -1)

	// Remove points from owner
	s.removeReactionPoints(targetType, targetID, reactionType)

	return nil
}

func (s *reactionService) GetUserReaction(userID uint, targetType models.ReactionTargetType, targetID uint) (*models.Reaction, error) {
	return s.reactionRepo.FindByUserAndTarget(userID, targetType, targetID)
}

// updateTargetCounts updates denormalized counts on report, comment, or forum post
func (s *reactionService) updateTargetCounts(targetType models.ReactionTargetType, targetID uint, reactionType models.ReactionType, delta int) {
	switch targetType {
	case models.ReactionTargetReport:
		report, err := s.reportRepo.FindByID(targetID)
		if err == nil {
			if reactionType == models.ReactionUpvote {
				report.Upvotes += delta
			} else {
				report.Downvotes += delta
			}
			s.reportRepo.Update(report)
		}
	case models.ReactionTargetComment:
		comment, err := s.commentRepo.FindByID(targetID)
		if err == nil {
			if reactionType == models.ReactionUpvote {
				comment.Upvotes += delta
			} else {
				comment.Downvotes += delta
			}
			s.commentRepo.Update(comment)
		}
	case models.ReactionTargetForumPost:
		if s.forumPostRepo != nil {
			forumPost, err := s.forumPostRepo.FindByID(targetID)
			if err == nil {
				if reactionType == models.ReactionUpvote {
					forumPost.Upvotes += delta
				} else {
					forumPost.Downvotes += delta
				}
				s.forumPostRepo.Update(forumPost)
			}
		}
	}
}

// awardReactionPoints gives points to the owner when their content is reacted to
func (s *reactionService) awardReactionPoints(targetType models.ReactionTargetType, targetID uint, reactionType models.ReactionType) {
	var ownerID uint

	switch targetType {
	case models.ReactionTargetReport:
		report, err := s.reportRepo.FindByID(targetID)
		if err != nil {
			return
		}
		ownerID = report.UserID
	case models.ReactionTargetComment:
		comment, err := s.commentRepo.FindByID(targetID)
		if err != nil {
			return
		}
		ownerID = comment.UserID
	case models.ReactionTargetForumPost:
		if s.forumPostRepo == nil {
			return
		}
		forumPost, err := s.forumPostRepo.FindByID(targetID)
		if err != nil {
			return
		}
		ownerID = forumPost.UserID
	default:
		return
	}

	// Award points based on reaction type
	if reactionType == models.ReactionUpvote {
		if targetType == models.ReactionTargetReport {
			s.reputationService.AddPoints(ownerID, 1) // Report upvote = +1
		} else if targetType == models.ReactionTargetComment {
			s.reputationService.AddPoints(ownerID, 2) // Comment upvote = +2
		} else if targetType == models.ReactionTargetForumPost {
			s.reputationService.AddPoints(ownerID, 1) // Forum post upvote = +1
		}
	} else {
		// Downvote removes points
		if targetType == models.ReactionTargetReport {
			s.reputationService.AddPoints(ownerID, -1) // Report downvote = -1
		} else if targetType == models.ReactionTargetComment {
			s.reputationService.AddPoints(ownerID, -1) // Comment downvote = -1
		} else if targetType == models.ReactionTargetForumPost {
			s.reputationService.AddPoints(ownerID, -1) // Forum post downvote = -1
		}
	}
}

// removeReactionPoints removes points when reaction is removed
func (s *reactionService) removeReactionPoints(targetType models.ReactionTargetType, targetID uint, reactionType models.ReactionType) {
	var ownerID uint

	switch targetType {
	case models.ReactionTargetReport:
		report, err := s.reportRepo.FindByID(targetID)
		if err != nil {
			return
		}
		ownerID = report.UserID
	case models.ReactionTargetComment:
		comment, err := s.commentRepo.FindByID(targetID)
		if err != nil {
			return
		}
		ownerID = comment.UserID
	case models.ReactionTargetForumPost:
		if s.forumPostRepo == nil {
			return
		}
		forumPost, err := s.forumPostRepo.FindByID(targetID)
		if err != nil {
			return
		}
		ownerID = forumPost.UserID
	default:
		return
	}

	// Remove points (reverse the award)
	if reactionType == models.ReactionUpvote {
		if targetType == models.ReactionTargetReport {
			s.reputationService.AddPoints(ownerID, -1)
		} else if targetType == models.ReactionTargetComment {
			s.reputationService.AddPoints(ownerID, -2)
		} else if targetType == models.ReactionTargetForumPost {
			s.reputationService.AddPoints(ownerID, -1)
		}
	} else {
		if targetType == models.ReactionTargetReport {
			s.reputationService.AddPoints(ownerID, 1)
		} else if targetType == models.ReactionTargetComment {
			s.reputationService.AddPoints(ownerID, 1)
		} else if targetType == models.ReactionTargetForumPost {
			s.reputationService.AddPoints(ownerID, 1)
		}
	}
}

// updateOwnerReputation handles reaction type change
func (s *reactionService) updateOwnerReputation(targetType models.ReactionTargetType, targetID uint, oldType, newType models.ReactionType) {
	var ownerID uint

	switch targetType {
	case models.ReactionTargetReport:
		report, err := s.reportRepo.FindByID(targetID)
		if err != nil {
			return
		}
		ownerID = report.UserID
	case models.ReactionTargetComment:
		comment, err := s.commentRepo.FindByID(targetID)
		if err != nil {
			return
		}
		ownerID = comment.UserID
	case models.ReactionTargetForumPost:
		if s.forumPostRepo == nil {
			return
		}
		forumPost, err := s.forumPostRepo.FindByID(targetID)
		if err != nil {
			return
		}
		ownerID = forumPost.UserID
	default:
		return
	}

	// Remove old reaction points
	if oldType == models.ReactionUpvote {
		if targetType == models.ReactionTargetReport {
			s.reputationService.AddPoints(ownerID, -1)
		} else if targetType == models.ReactionTargetComment {
			s.reputationService.AddPoints(ownerID, -2)
		} else if targetType == models.ReactionTargetForumPost {
			s.reputationService.AddPoints(ownerID, -1)
		}
	} else {
		if targetType == models.ReactionTargetReport {
			s.reputationService.AddPoints(ownerID, 1)
		} else if targetType == models.ReactionTargetComment {
			s.reputationService.AddPoints(ownerID, 1)
		} else if targetType == models.ReactionTargetForumPost {
			s.reputationService.AddPoints(ownerID, 1)
		}
	}

	// Add new reaction points
	if newType == models.ReactionUpvote {
		if targetType == models.ReactionTargetReport {
			s.reputationService.AddPoints(ownerID, 1)
		} else if targetType == models.ReactionTargetComment {
			s.reputationService.AddPoints(ownerID, 2)
		} else if targetType == models.ReactionTargetForumPost {
			s.reputationService.AddPoints(ownerID, 1)
		}
	} else {
		if targetType == models.ReactionTargetReport {
			s.reputationService.AddPoints(ownerID, -1)
		} else if targetType == models.ReactionTargetComment {
			s.reputationService.AddPoints(ownerID, -1)
		} else if targetType == models.ReactionTargetForumPost {
			s.reputationService.AddPoints(ownerID, -1)
		}
	}
}

