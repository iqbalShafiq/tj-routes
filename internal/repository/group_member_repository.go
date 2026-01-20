package repository

import (
	"time"

	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type GroupMemberRepository interface {
	Create(member *models.GroupMember) error
	FindByID(id uint) (*models.GroupMember, error)
	FindByGroupAndUser(groupID, userID uint) (*models.GroupMember, error)
	ListByGroupID(groupID uint, offset, limit int) ([]models.GroupMember, int64, error)
	ListByUserID(userID uint, offset, limit int) ([]models.GroupMember, int64, error)
	UpdateRole(groupID, userID uint, role string) error
	UpdateLastRead(groupID, userID uint, lastReadAt time.Time) error
	UpdateMutedUntil(groupID, userID uint, mutedUntil *time.Time) error
	Delete(groupID, userID uint) error
	CountByGroupID(groupID uint) (int64, error)
}

type groupMemberRepository struct {
	db *gorm.DB
}

func NewGroupMemberRepository(db *gorm.DB) GroupMemberRepository {
	return &groupMemberRepository{db: db}
}

func (r *groupMemberRepository) Create(member *models.GroupMember) error {
	return r.db.Create(member).Error
}

func (r *groupMemberRepository) FindByID(id uint) (*models.GroupMember, error) {
	var member models.GroupMember
	err := r.db.Preload("Group").
		Preload("User").
		First(&member, id).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *groupMemberRepository) FindByGroupAndUser(groupID, userID uint) (*models.GroupMember, error) {
	var member models.GroupMember
	err := r.db.Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *groupMemberRepository) ListByGroupID(groupID uint, offset, limit int) ([]models.GroupMember, int64, error) {
	var members []models.GroupMember
	var total int64

	query := r.db.Model(&models.GroupMember{}).Where("group_id = ?", groupID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("User").
		Order("role ASC, joined_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&members).Error

	return members, total, err
}

func (r *groupMemberRepository) ListByUserID(userID uint, offset, limit int) ([]models.GroupMember, int64, error) {
	var members []models.GroupMember
	var total int64

	query := r.db.Model(&models.GroupMember{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Group").
		Order("joined_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&members).Error

	return members, total, err
}

func (r *groupMemberRepository) UpdateRole(groupID, userID uint, role string) error {
	return r.db.Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Update("role", role).Error
}

func (r *groupMemberRepository) UpdateLastRead(groupID, userID uint, lastReadAt time.Time) error {
	return r.db.Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Update("last_read_at", lastReadAt).Error
}

func (r *groupMemberRepository) UpdateMutedUntil(groupID, userID uint, mutedUntil *time.Time) error {
	return r.db.Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Update("muted_until", mutedUntil).Error
}

func (r *groupMemberRepository) Delete(groupID, userID uint) error {
	return r.db.Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&models.GroupMember{}).Error
}

func (r *groupMemberRepository) CountByGroupID(groupID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.GroupMember{}).
		Where("group_id = ?", groupID).
		Count(&count).Error
	return count, err
}
