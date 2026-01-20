package repository

import (
	"time"

	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type GroupInviteRepository interface {
	Create(invite *models.GroupInvite) error
	FindByID(id uint) (*models.GroupInvite, error)
	FindByGroupAndInvitee(groupID, inviteeID uint) (*models.GroupInvite, error)
	ListByGroupID(groupID uint, offset, limit int) ([]models.GroupInvite, int64, error)
	ListByInviteeID(inviteeID uint, offset, limit int) ([]models.GroupInvite, int64, error)
	UpdateStatus(id uint, status string) error
	Delete(id uint) error
	MarkExpired() error
}

type groupInviteRepository struct {
	db *gorm.DB
}

func NewGroupInviteRepository(db *gorm.DB) GroupInviteRepository {
	return &groupInviteRepository{db: db}
}

func (r *groupInviteRepository) Create(invite *models.GroupInvite) error {
	return r.db.Create(invite).Error
}

func (r *groupInviteRepository) FindByID(id uint) (*models.GroupInvite, error) {
	var invite models.GroupInvite
	err := r.db.Preload("Group").
		Preload("Inviter").
		Preload("Invitee").
		First(&invite, id).Error
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *groupInviteRepository) FindByGroupAndInvitee(groupID, inviteeID uint) (*models.GroupInvite, error) {
	var invite models.GroupInvite
	err := r.db.Where("group_id = ? AND invitee_id = ?", groupID, inviteeID).
		Order("created_at DESC").
		First(&invite).Error
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *groupInviteRepository) ListByGroupID(groupID uint, offset, limit int) ([]models.GroupInvite, int64, error) {
	var invites []models.GroupInvite
	var total int64

	query := r.db.Model(&models.GroupInvite{}).Where("group_id = ?", groupID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Invitee").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&invites).Error

	return invites, total, err
}

func (r *groupInviteRepository) ListByInviteeID(inviteeID uint, offset, limit int) ([]models.GroupInvite, int64, error) {
	var invites []models.GroupInvite
	var total int64

	query := r.db.Model(&models.GroupInvite{}).Where("invitee_id = ?", inviteeID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Group").
		Preload("Inviter").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&invites).Error

	return invites, total, err
}

func (r *groupInviteRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.GroupInvite{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *groupInviteRepository) Delete(id uint) error {
	return r.db.Delete(&models.GroupInvite{}, id).Error
}

func (r *groupInviteRepository) MarkExpired() error {
	now := time.Now()
	return r.db.Model(&models.GroupInvite{}).
		Where("status = ? AND expires_at < ?", "pending", now).
		Update("status", "expired").Error
}
