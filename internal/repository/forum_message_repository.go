package repository

import (
	"time"

	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type ForumMessageRepository interface {
	Create(message *models.ForumMessage) error
	FindByID(id uint) (*models.ForumMessage, error)
	ListByForumID(forumID uint, offset, limit int) ([]models.ForumMessage, int64, error)
	Delete(id uint) error
	DeleteOlderThan(cutoff time.Time) error
}

type forumMessageRepository struct {
	db *gorm.DB
}

func NewForumMessageRepository(db *gorm.DB) ForumMessageRepository {
	return &forumMessageRepository{db: db}
}

func (r *forumMessageRepository) Create(message *models.ForumMessage) error {
	return r.db.Create(message).Error
}

func (r *forumMessageRepository) FindByID(id uint) (*models.ForumMessage, error) {
	var message models.ForumMessage
	err := r.db.Preload("Forum").
		Preload("User").
		First(&message, id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *forumMessageRepository) ListByForumID(forumID uint, offset, limit int) ([]models.ForumMessage, int64, error) {
	var messages []models.ForumMessage
	var total int64

	query := r.db.Model(&models.ForumMessage{}).Where("forum_id = ?", forumID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("User").
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error

	return messages, total, err
}

func (r *forumMessageRepository) Delete(id uint) error {
	return r.db.Delete(&models.ForumMessage{}, id).Error
}

func (r *forumMessageRepository) DeleteOlderThan(cutoff time.Time) error {
	return r.db.Where("created_at < ?", cutoff).
		Delete(&models.ForumMessage{}).Error
}
