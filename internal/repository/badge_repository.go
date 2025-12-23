package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type BadgeRepository interface {
	Create(badge *models.Badge) error
	FindByID(id uint) (*models.Badge, error)
	FindAll() ([]models.Badge, error)
	FindByCriteriaType(criteriaType models.BadgeCriteriaType) ([]models.Badge, error)
	Update(badge *models.Badge) error
	Delete(id uint) error
}

type UserBadgeRepository interface {
	Create(userBadge *models.UserBadge) error
	FindByUserID(userID uint) ([]models.UserBadge, error)
	FindByUserAndBadge(userID, badgeID uint) (*models.UserBadge, error)
	FindByBadgeID(badgeID uint) ([]models.UserBadge, error)
	Delete(userID, badgeID uint) error
}

type badgeRepository struct {
	db *gorm.DB
}

func NewBadgeRepository(db *gorm.DB) BadgeRepository {
	return &badgeRepository{db: db}
}

func (r *badgeRepository) Create(badge *models.Badge) error {
	return r.db.Create(badge).Error
}

func (r *badgeRepository) FindByID(id uint) (*models.Badge, error) {
	var badge models.Badge
	err := r.db.First(&badge, id).Error
	if err != nil {
		return nil, err
	}
	return &badge, nil
}

func (r *badgeRepository) FindAll() ([]models.Badge, error) {
	var badges []models.Badge
	err := r.db.Order("criteria_value ASC").Find(&badges).Error
	return badges, err
}

func (r *badgeRepository) FindByCriteriaType(criteriaType models.BadgeCriteriaType) ([]models.Badge, error) {
	var badges []models.Badge
	err := r.db.Where("criteria_type = ?", criteriaType).
		Order("criteria_value ASC").
		Find(&badges).Error
	return badges, err
}

func (r *badgeRepository) Update(badge *models.Badge) error {
	return r.db.Save(badge).Error
}

func (r *badgeRepository) Delete(id uint) error {
	return r.db.Delete(&models.Badge{}, id).Error
}

type userBadgeRepository struct {
	db *gorm.DB
}

func NewUserBadgeRepository(db *gorm.DB) UserBadgeRepository {
	return &userBadgeRepository{db: db}
}

func (r *userBadgeRepository) Create(userBadge *models.UserBadge) error {
	return r.db.Create(userBadge).Error
}

func (r *userBadgeRepository) FindByUserID(userID uint) ([]models.UserBadge, error) {
	var userBadges []models.UserBadge
	err := r.db.Where("user_id = ?", userID).
		Preload("Badge").
		Order("earned_at DESC").
		Find(&userBadges).Error
	return userBadges, err
}

func (r *userBadgeRepository) FindByUserAndBadge(userID, badgeID uint) (*models.UserBadge, error) {
	var userBadge models.UserBadge
	err := r.db.Where("user_id = ? AND badge_id = ?", userID, badgeID).
		First(&userBadge).Error
	if err != nil {
		return nil, err
	}
	return &userBadge, nil
}

func (r *userBadgeRepository) FindByBadgeID(badgeID uint) ([]models.UserBadge, error) {
	var userBadges []models.UserBadge
	err := r.db.Where("badge_id = ?", badgeID).
		Preload("User").
		Find(&userBadges).Error
	return userBadges, err
}

func (r *userBadgeRepository) Delete(userID, badgeID uint) error {
	return r.db.Where("user_id = ? AND badge_id = ?", userID, badgeID).
		Delete(&models.UserBadge{}).Error
}

