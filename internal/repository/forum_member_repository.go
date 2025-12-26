package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type ForumMemberRepository interface {
	Create(member *models.ForumMember) error
	FindByForumAndUser(forumID uint, userID uint) (*models.ForumMember, error)
	Delete(forumID uint, userID uint) error
	ListMembers(forumID uint, offset, limit int) ([]models.ForumMember, int64, error)
	ListForumsByUser(userID uint, offset, limit int) ([]models.Forum, int64, error)
	CountByForumID(forumID uint) (int64, error)
	IsMember(forumID uint, userID uint) (bool, error)
}

type forumMemberRepository struct {
	db *gorm.DB
}

func NewForumMemberRepository(db *gorm.DB) ForumMemberRepository {
	return &forumMemberRepository{db: db}
}

func (r *forumMemberRepository) Create(member *models.ForumMember) error {
	return r.db.Create(member).Error
}

func (r *forumMemberRepository) FindByForumAndUser(forumID uint, userID uint) (*models.ForumMember, error) {
	var member models.ForumMember
	err := r.db.Where("forum_id = ? AND user_id = ?", forumID, userID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *forumMemberRepository) Delete(forumID uint, userID uint) error {
	return r.db.Where("forum_id = ? AND user_id = ?", forumID, userID).
		Delete(&models.ForumMember{}).Error
}

func (r *forumMemberRepository) ListMembers(forumID uint, offset, limit int) ([]models.ForumMember, int64, error) {
	var members []models.ForumMember
	var total int64

	query := r.db.Model(&models.ForumMember{}).Where("forum_id = ?", forumID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&members).Error

	return members, total, err
}

func (r *forumMemberRepository) ListForumsByUser(userID uint, offset, limit int) ([]models.Forum, int64, error) {
	var forums []models.Forum
	var total int64

	// Get forum IDs from memberships
	var memberIDs []uint
	if err := r.db.Model(&models.ForumMember{}).
		Where("user_id = ?", userID).
		Pluck("forum_id", &memberIDs).Error; err != nil {
		return nil, 0, err
	}

	if len(memberIDs) == 0 {
		return []models.Forum{}, 0, nil
	}

	query := r.db.Model(&models.Forum{}).Where("id IN ?", memberIDs)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Route").
		Offset(offset).
		Limit(limit).
		Find(&forums).Error

	return forums, total, err
}

func (r *forumMemberRepository) CountByForumID(forumID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.ForumMember{}).Where("forum_id = ?", forumID).Count(&count).Error
	return count, err
}

func (r *forumMemberRepository) IsMember(forumID uint, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.ForumMember{}).
		Where("forum_id = ? AND user_id = ?", forumID, userID).
		Count(&count).Error
	return count > 0, err
}

