package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type ReactionRepository interface {
	Create(reaction *models.Reaction) error
	Update(reaction *models.Reaction) error
	FindByUserAndTarget(userID uint, targetType models.ReactionTargetType, targetID uint) (*models.Reaction, error)
	FindByTarget(targetType models.ReactionTargetType, targetID uint) ([]models.Reaction, error)
	CountByTargetAndType(targetType models.ReactionTargetType, targetID uint, reactionType models.ReactionType) (int64, error)
	CountUpvotesByUserContent(userID uint) (int64, error)
	Delete(reactionID uint) error
	DeleteByUserAndTarget(userID uint, targetType models.ReactionTargetType, targetID uint) error
}

type reactionRepository struct {
	db *gorm.DB
}

func NewReactionRepository(db *gorm.DB) ReactionRepository {
	return &reactionRepository{db: db}
}

func (r *reactionRepository) Create(reaction *models.Reaction) error {
	return r.db.Create(reaction).Error
}

func (r *reactionRepository) Update(reaction *models.Reaction) error {
	return r.db.Save(reaction).Error
}

func (r *reactionRepository) FindByUserAndTarget(userID uint, targetType models.ReactionTargetType, targetID uint) (*models.Reaction, error) {
	var reaction models.Reaction
	err := r.db.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		First(&reaction).Error
	if err != nil {
		return nil, err
	}
	return &reaction, nil
}

func (r *reactionRepository) FindByTarget(targetType models.ReactionTargetType, targetID uint) ([]models.Reaction, error) {
	var reactions []models.Reaction
	err := r.db.Where("target_type = ? AND target_id = ?", targetType, targetID).
		Preload("User").
		Find(&reactions).Error
	return reactions, err
}

func (r *reactionRepository) CountByTargetAndType(targetType models.ReactionTargetType, targetID uint, reactionType models.ReactionType) (int64, error) {
	var count int64
	err := r.db.Model(&models.Reaction{}).
		Where("target_type = ? AND target_id = ? AND reaction_type = ?", targetType, targetID, reactionType).
		Count(&count).Error
	return count, err
}

func (r *reactionRepository) Delete(reactionID uint) error {
	return r.db.Delete(&models.Reaction{}, reactionID).Error
}

func (r *reactionRepository) DeleteByUserAndTarget(userID uint, targetType models.ReactionTargetType, targetID uint) error {
	return r.db.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		Delete(&models.Reaction{}).Error
}

func (r *reactionRepository) CountUpvotesByUserContent(userID uint) (int64, error) {
	var reportUpvotes int64
	// Count upvotes on reports by user
	err := r.db.Table("reactions").
		Joins("JOIN reports ON reactions.target_id = reports.id").
		Where("reactions.target_type = ? AND reactions.reaction_type = ? AND reports.user_id = ?",
			models.ReactionTargetReport, models.ReactionUpvote, userID).
		Count(&reportUpvotes).Error
	if err != nil {
		return 0, err
	}

	// Count upvotes on comments by user
	var commentUpvotes int64
	err = r.db.Table("reactions").
		Joins("JOIN comments ON reactions.target_id = comments.id").
		Where("reactions.target_type = ? AND reactions.reaction_type = ? AND comments.user_id = ?",
			models.ReactionTargetComment, models.ReactionUpvote, userID).
		Count(&commentUpvotes).Error
	if err != nil {
		return 0, err
	}

	return reportUpvotes + commentUpvotes, nil
}

