package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(message *models.Message) error
	FindByID(id uint) (*models.Message, error)
	ListByConversationID(conversationID uint, offset, limit int) ([]models.Message, int64, error)
	ListByGroupID(groupID uint, offset, limit int) ([]models.Message, int64, error)
	Update(message *models.Message) error
	Delete(id uint) error
	UpdateStatus(id uint, status string) error
	CountUnreadByConversation(conversationID, userID uint) (int64, error)
	CountUnreadByGroup(groupID, userID uint) (int64, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(message *models.Message) error {
	return r.db.Create(message).Error
}

func (r *messageRepository) FindByID(id uint) (*models.Message, error) {
	var message models.Message
	err := r.db.Preload("Sender").
		Preload("ReplyTo").
		Preload("Reactions.User").
		First(&message, id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) ListByConversationID(conversationID uint, offset, limit int) ([]models.Message, int64, error) {
	var messages []models.Message
	var total int64

	query := r.db.Model(&models.Message{}).Where("conversation_id = ?", conversationID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Sender").
		Preload("Reactions.User").
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error

	return messages, total, err
}

func (r *messageRepository) ListByGroupID(groupID uint, offset, limit int) ([]models.Message, int64, error) {
	var messages []models.Message
	var total int64

	query := r.db.Model(&models.Message{}).Where("group_id = ?", groupID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Sender").
		Preload("Reactions.User").
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error

	return messages, total, err
}

func (r *messageRepository) Update(message *models.Message) error {
	return r.db.Save(message).Error
}

func (r *messageRepository) Delete(id uint) error {
	return r.db.Delete(&models.Message{}, id).Error
}

func (r *messageRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.Message{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *messageRepository) CountUnreadByConversation(conversationID, userID uint) (int64, error) {
	var count int64

	subQuery := r.db.Model(&models.ConversationParticipant{}).
		Select("last_read_at").
		Where("conversation_id = ? AND user_id = ?", conversationID, userID)

	err := r.db.Model(&models.Message{}).
		Where("conversation_id = ? AND sender_id != ? AND created_at > (?)", conversationID, userID, subQuery).
		Count(&count).Error

	return count, err
}

func (r *messageRepository) CountUnreadByGroup(groupID, userID uint) (int64, error) {
	var count int64

	subQuery := r.db.Model(&models.GroupMember{}).
		Select("last_read_at").
		Where("group_id = ? AND user_id = ?", groupID, userID)

	err := r.db.Model(&models.Message{}).
		Where("group_id = ? AND sender_id != ? AND created_at > (?)", groupID, userID, subQuery).
		Count(&count).Error

	return count, err
}
