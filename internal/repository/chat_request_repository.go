package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type ChatRequestRepository interface {
	Create(request *models.ChatRequest) error
	FindByID(id uint) (*models.ChatRequest, error)
	FindBySenderAndReceiver(senderID, receiverID uint) (*models.ChatRequest, error)
	ListSent(senderID uint, offset, limit int) ([]models.ChatRequest, int64, error)
	ListReceived(receiverID uint, offset, limit int) ([]models.ChatRequest, int64, error)
	UpdateStatus(id uint, status string) error
	Delete(id uint) error
}

type chatRequestRepository struct {
	db *gorm.DB
}

func NewChatRequestRepository(db *gorm.DB) ChatRequestRepository {
	return &chatRequestRepository{db: db}
}

func (r *chatRequestRepository) Create(request *models.ChatRequest) error {
	return r.db.Create(request).Error
}

func (r *chatRequestRepository) FindByID(id uint) (*models.ChatRequest, error) {
	var request models.ChatRequest
	err := r.db.Preload("Sender").
		Preload("Receiver").
		First(&request, id).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *chatRequestRepository) FindBySenderAndReceiver(senderID, receiverID uint) (*models.ChatRequest, error) {
	var request models.ChatRequest
	err := r.db.Where("sender_id = ? AND receiver_id = ?", senderID, receiverID).
		First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *chatRequestRepository) ListSent(senderID uint, offset, limit int) ([]models.ChatRequest, int64, error) {
	var requests []models.ChatRequest
	var total int64

	query := r.db.Model(&models.ChatRequest{}).
		Where("sender_id = ?", senderID).
		Where("status = ?", "pending")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Receiver").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&requests).Error

	return requests, total, err
}

func (r *chatRequestRepository) ListReceived(receiverID uint, offset, limit int) ([]models.ChatRequest, int64, error) {
	var requests []models.ChatRequest
	var total int64

	query := r.db.Model(&models.ChatRequest{}).
		Where("receiver_id = ?", receiverID).
		Where("status = ?", "pending")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Sender").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&requests).Error

	return requests, total, err
}

func (r *chatRequestRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.ChatRequest{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *chatRequestRepository) Delete(id uint) error {
	return r.db.Delete(&models.ChatRequest{}, id).Error
}
