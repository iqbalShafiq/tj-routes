package repository

import (
	"tj-routes/internal/models"
	"gorm.io/gorm"
)

// UserPlaceRepository interface for user place operations
type UserPlaceRepository interface {
	Create(place *models.UserPlace) error
	Update(place *models.UserPlace) error
	Delete(id uint) error
	FindByID(id uint) (*models.UserPlace, error)
	FindByUserID(userID uint) ([]models.UserPlace, error)
	FindByUserIDAndType(userID uint, placeType models.PlaceType) ([]models.UserPlace, error)
	FindDefaultByUserID(userID uint, placeType models.PlaceType) (*models.UserPlace, error)
	SetDefault(userID uint, placeID uint, placeType models.PlaceType) error
	CountByUserID(userID uint) (int64, error)
}

type userPlaceRepository struct {
	db *gorm.DB
}

// NewUserPlaceRepository creates a new UserPlaceRepository
func NewUserPlaceRepository(db *gorm.DB) UserPlaceRepository {
	return &userPlaceRepository{db: db}
}

// Create creates a new user place
func (r *userPlaceRepository) Create(place *models.UserPlace) error {
	return r.db.Create(place).Error
}

// Update updates an existing user place
func (r *userPlaceRepository) Update(place *models.UserPlace) error {
	return r.db.Save(place).Error
}

// Delete deletes a user place
func (r *userPlaceRepository) Delete(id uint) error {
	return r.db.Delete(&models.UserPlace{}, id).Error
}

// FindByID finds a user place by ID
func (r *userPlaceRepository) FindByID(id uint) (*models.UserPlace, error) {
	var place models.UserPlace
	err := r.db.First(&place, id).Error
	if err != nil {
		return nil, err
	}
	return &place, nil
}

// FindByUserID finds all places for a user
func (r *userPlaceRepository) FindByUserID(userID uint) ([]models.UserPlace, error) {
	var places []models.UserPlace
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&places).Error
	return places, err
}

// FindByUserIDAndType finds places by user ID and place type
func (r *userPlaceRepository) FindByUserIDAndType(userID uint, placeType models.PlaceType) ([]models.UserPlace, error) {
	var places []models.UserPlace
	err := r.db.Where("user_id = ? AND place_type = ?", userID, placeType).Find(&places).Error
	return places, err
}

// FindDefaultByUserID finds the default place of a specific type for a user
func (r *userPlaceRepository) FindDefaultByUserID(userID uint, placeType models.PlaceType) (*models.UserPlace, error) {
	var place models.UserPlace
	err := r.db.Where("user_id = ? AND place_type = ? AND is_default = ?", userID, placeType, true).First(&place).Error
	if err != nil {
		return nil, err
	}
	return &place, nil
}

// SetDefault sets a place as default and unsets other places of the same type
func (r *userPlaceRepository) SetDefault(userID uint, placeID uint, placeType models.PlaceType) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Unset all defaults for this place type
		if err := tx.Model(&models.UserPlace{}).
			Where("user_id = ? AND place_type = ?", userID, placeType).
			Update("is_default", false).Error; err != nil {
			return err
		}
		// Set the new default
		return tx.Model(&models.UserPlace{}).
			Where("id = ? AND user_id = ?", placeID, userID).
			Update("is_default", true).Error
	})
}

// CountByUserID counts places by user ID
func (r *userPlaceRepository) CountByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.UserPlace{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
