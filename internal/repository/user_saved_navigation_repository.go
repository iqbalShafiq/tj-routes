package repository

import (
	"tj-routes/internal/models"
	"gorm.io/gorm"
)

// UserSavedNavigationRepository interface for user saved navigation operations
type UserSavedNavigationRepository interface {
	Create(nav *models.UserSavedNavigation) error
	Update(nav *models.UserSavedNavigation) error
	Delete(id uint) error
	FindByID(id uint) (*models.UserSavedNavigation, error)
	FindByUserID(userID uint, offset, limit int) ([]models.UserSavedNavigation, int64, error)
	CountByUserID(userID uint) (int64, error)
}

type userSavedNavigationRepository struct {
	db *gorm.DB
}

// NewUserSavedNavigationRepository creates a new UserSavedNavigationRepository
func NewUserSavedNavigationRepository(db *gorm.DB) UserSavedNavigationRepository {
	return &userSavedNavigationRepository{db: db}
}

// Create creates a new user saved navigation
func (r *userSavedNavigationRepository) Create(nav *models.UserSavedNavigation) error {
	return r.db.Create(nav).Error
}

// Update updates an existing user saved navigation
func (r *userSavedNavigationRepository) Update(nav *models.UserSavedNavigation) error {
	return r.db.Save(nav).Error
}

// Delete deletes a user saved navigation
func (r *userSavedNavigationRepository) Delete(id uint) error {
	return r.db.Delete(&models.UserSavedNavigation{}, id).Error
}

// FindByID finds a user saved navigation by ID
func (r *userSavedNavigationRepository) FindByID(id uint) (*models.UserSavedNavigation, error) {
	var nav models.UserSavedNavigation
	err := r.db.Preload("FromPlace").Preload("FromStop").Preload("ToPlace").Preload("ToStop").
		First(&nav, id).Error
	if err != nil {
		return nil, err
	}
	return &nav, nil
}

// FindByUserID finds all saved navigations by user ID with pagination
func (r *userSavedNavigationRepository) FindByUserID(userID uint, offset, limit int) ([]models.UserSavedNavigation, int64, error) {
	var navigations []models.UserSavedNavigation
	var total int64

	query := r.db.Model(&models.UserSavedNavigation{}).Where("user_id = ?", userID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with preloads
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").
		Preload("FromPlace").
		Preload("FromStop").
		Preload("ToPlace").
		Preload("ToStop").
		Find(&navigations).Error

	return navigations, total, err
}

// CountByUserID counts saved navigations by user ID
func (r *userSavedNavigationRepository) CountByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.UserSavedNavigation{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
