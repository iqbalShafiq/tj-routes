package repository

import (
	"time"

	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type ConversationRepository interface {
	Create(conversation *models.Conversation) error
	FindByID(id uint) (*models.Conversation, error)
	FindByUsers(userID1, userID2 uint) (*models.Conversation, error)
	ListByUserID(userID uint, offset, limit int) ([]models.Conversation, int64, error)
	Update(conversation *models.Conversation) error
	Delete(id uint) error
	AddParticipant(participant *models.ConversationParticipant) error
	RemoveParticipant(conversationID, userID uint) error
	FindParticipants(conversationID uint) ([]models.ConversationParticipant, error)
	UpdateLastRead(conversationID, userID uint, lastReadAt time.Time) error
}

type conversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) Create(conversation *models.Conversation) error {
	return r.db.Create(conversation).Error
}

func (r *conversationRepository) FindByID(id uint) (*models.Conversation, error) {
	var conversation models.Conversation
	err := r.db.Preload("Participants.User").
		Preload("Messages.Sender").
		First(&conversation, id).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (r *conversationRepository) FindByUsers(userID1, userID2 uint) (*models.Conversation, error) {
	var conversation models.Conversation

	query := `
		SELECT c.* FROM conversations c
		INNER JOIN conversation_participants cp1 ON c.id = cp1.conversation_id
		INNER JOIN conversation_participants cp2 ON c.id = cp2.conversation_id
		WHERE cp1.user_id = ? AND cp2.user_id = ? AND c.type = 'direct'
		LIMIT 1
	`

	err := r.db.Raw(query, userID1, userID2).First(&conversation).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (r *conversationRepository) ListByUserID(userID uint, offset, limit int) ([]models.Conversation, int64, error) {
	var conversations []models.Conversation
	var total int64

	query := r.db.Model(&models.Conversation{}).
		Joins("INNER JOIN conversation_participants ON conversations.id = conversation_participants.conversation_id").
		Where("conversation_participants.user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Participants.User").
		Order("conversations.updated_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&conversations).Error

	return conversations, total, err
}

func (r *conversationRepository) Update(conversation *models.Conversation) error {
	return r.db.Save(conversation).Error
}

func (r *conversationRepository) Delete(id uint) error {
	return r.db.Delete(&models.Conversation{}, id).Error
}

func (r *conversationRepository) AddParticipant(participant *models.ConversationParticipant) error {
	return r.db.Create(participant).Error
}

func (r *conversationRepository) RemoveParticipant(conversationID, userID uint) error {
	return r.db.Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Delete(&models.ConversationParticipant{}).Error
}

func (r *conversationRepository) FindParticipants(conversationID uint) ([]models.ConversationParticipant, error) {
	var participants []models.ConversationParticipant
	err := r.db.Preload("User").
		Where("conversation_id = ?", conversationID).
		Find(&participants).Error
	return participants, err
}

func (r *conversationRepository) UpdateLastRead(conversationID, userID uint, lastReadAt time.Time) error {
	return r.db.Model(&models.ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("last_read_at", lastReadAt).Error
}
