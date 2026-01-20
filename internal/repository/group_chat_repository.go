package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type GroupChatRepository interface {
	Create(group *models.GroupChat) error
	FindByID(id uint) (*models.GroupChat, error)
	ListByUserID(userID uint, offset, limit int) ([]models.GroupChat, int64, error)
	Update(group *models.GroupChat) error
	Delete(id uint) error
	UpdateAvatar(groupID uint, avatarURL string) error
}

type groupChatRepository struct {
	db *gorm.DB
}

func NewGroupChatRepository(db *gorm.DB) GroupChatRepository {
	return &groupChatRepository{db: db}
}

func (r *groupChatRepository) Create(group *models.GroupChat) error {
	return r.db.Create(group).Error
}

func (r *groupChatRepository) FindByID(id uint) (*models.GroupChat, error) {
	var group models.GroupChat
	err := r.db.Preload("CreatedBy").
		Preload("Members.User").
		First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *groupChatRepository) ListByUserID(userID uint, offset, limit int) ([]models.GroupChat, int64, error) {
	var groups []models.GroupChat
	var total int64

	query := r.db.Model(&models.GroupChat{}).
		Joins("INNER JOIN group_members ON group_chats.id = group_members.group_id").
		Where("group_members.user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("CreatedBy").
		Order("group_chats.updated_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&groups).Error

	return groups, total, err
}

func (r *groupChatRepository) Update(group *models.GroupChat) error {
	return r.db.Save(group).Error
}

func (r *groupChatRepository) Delete(id uint) error {
	return r.db.Delete(&models.GroupChat{}, id).Error
}

func (r *groupChatRepository) UpdateAvatar(groupID uint, avatarURL string) error {
	return r.db.Model(&models.GroupChat{}).
		Where("id = ?", groupID).
		Update("avatar_url", avatarURL).Error
}
