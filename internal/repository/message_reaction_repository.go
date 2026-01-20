package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type MessageReactionRepository interface {
	Create(reaction *models.MessageReaction) error
	FindByID(id uint) (*models.MessageReaction, error)
	FindByUserAndMessage(userID, messageID uint) (*models.MessageReaction, error)
	ListByMessageID(messageID uint) ([]models.MessageReaction, error)
	Delete(id uint) error
	DeleteByUserAndMessage(userID, messageID uint) error
}

type messageReactionRepository struct {
	db *gorm.DB
}

func NewMessageReactionRepository(db *gorm.DB) MessageReactionRepository {
	return &messageReactionRepository{db: db}
}

func (r *messageReactionRepository) Create(reaction *models.MessageReaction) error {
	return r.db.Create(reaction).Error
}

func (r *messageReactionRepository) FindByID(id uint) (*models.MessageReaction, error) {
	var reaction models.MessageReaction
	err := r.db.Preload("User").
		First(&reaction, id).Error
	if err != nil {
		return nil, err
	}
	return &reaction, nil
}

func (r *messageReactionRepository) FindByUserAndMessage(userID, messageID uint) (*models.MessageReaction, error) {
	var reaction models.MessageReaction
	err := r.db.Where("user_id = ? AND message_id = ?", userID, messageID).
		First(&reaction).Error
	if err != nil {
		return nil, err
	}
	return &reaction, nil
}

func (r *messageReactionRepository) ListByMessageID(messageID uint) ([]models.MessageReaction, error) {
	var reactions []models.MessageReaction
	err := r.db.Preload("User").
		Where("message_id = ?", messageID).
		Find(&reactions).Error
	return reactions, err
}

func (r *messageReactionRepository) Delete(id uint) error {
	return r.db.Delete(&models.MessageReaction{}, id).Error
}

func (r *messageReactionRepository) DeleteByUserAndMessage(userID, messageID uint) error {
	return r.db.Where("user_id = ? AND message_id = ?", userID, messageID).
		Delete(&models.MessageReaction{}).Error
}
