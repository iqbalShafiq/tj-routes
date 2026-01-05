package repository

import (
	"tj-routes/internal/models"
	"gorm.io/gorm"
)

// UserFavoriteRepository interface for user favorite operations
type UserFavoriteRepository interface {
	Create(favorite *models.UserFavorite) error
	Delete(userID uint, favoriteType models.FavoriteType, routeID, stopID *uint) error
	Exists(userID uint, favoriteType models.FavoriteType, routeID, stopID *uint) (bool, error)
	FindByUserID(userID uint, favoriteType models.FavoriteType, offset, limit int) ([]models.UserFavorite, int64, error)
	GetFavoriteRouteIDs(userID uint) ([]uint, error)
	GetFavoriteStopIDs(userID uint) ([]uint, error)
	CountByUserID(userID uint, favoriteType *models.FavoriteType) (int64, error)
}

type userFavoriteRepository struct {
	db *gorm.DB
}

// NewUserFavoriteRepository creates a new UserFavoriteRepository
func NewUserFavoriteRepository(db *gorm.DB) UserFavoriteRepository {
	return &userFavoriteRepository{db: db}
}

// Create creates a new user favorite
func (r *userFavoriteRepository) Create(favorite *models.UserFavorite) error {
	return r.db.Create(favorite).Error
}

// Delete deletes a user favorite
func (r *userFavoriteRepository) Delete(userID uint, favoriteType models.FavoriteType, routeID, stopID *uint) error {
	query := r.db.Where("user_id = ? AND favorite_type = ?", userID, favoriteType)
	if routeID != nil {
		query = query.Where("route_id = ?", *routeID)
	}
	if stopID != nil {
		query = query.Where("stop_id = ?", *stopID)
	}
	return query.Delete(&models.UserFavorite{}).Error
}

// Exists checks if a favorite exists
func (r *userFavoriteRepository) Exists(userID uint, favoriteType models.FavoriteType, routeID, stopID *uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.UserFavorite{}).
		Where("user_id = ? AND favorite_type = ?", userID, favoriteType)
	if routeID != nil {
		query = query.Where("route_id = ?", *routeID)
	}
	if stopID != nil {
		query = query.Where("stop_id = ?", *stopID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// FindByUserID finds all favorites by user ID with pagination
func (r *userFavoriteRepository) FindByUserID(userID uint, favoriteType models.FavoriteType, offset, limit int) ([]models.UserFavorite, int64, error) {
	var favorites []models.UserFavorite
	var total int64

	query := r.db.Model(&models.UserFavorite{}).Where("user_id = ? AND favorite_type = ?", userID, favoriteType)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with preloads
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").
		Preload("Route").
		Preload("Stop").
		Find(&favorites).Error

	return favorites, total, err
}

// GetFavoriteRouteIDs returns all favorite route IDs for a user
func (r *userFavoriteRepository) GetFavoriteRouteIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Model(&models.UserFavorite{}).
		Where("user_id = ? AND favorite_type = ? AND route_id IS NOT NULL", userID, models.FavoriteTypeRoute).
		Pluck("route_id", &ids).Error
	return ids, err
}

// GetFavoriteStopIDs returns all favorite stop IDs for a user
func (r *userFavoriteRepository) GetFavoriteStopIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Model(&models.UserFavorite{}).
		Where("user_id = ? AND favorite_type = ? AND stop_id IS NOT NULL", userID, models.FavoriteTypeStop).
		Pluck("stop_id", &ids).Error
	return ids, err
}

// CountByUserID counts favorites by user ID
func (r *userFavoriteRepository) CountByUserID(userID uint, favoriteType *models.FavoriteType) (int64, error) {
	var count int64
	query := r.db.Model(&models.UserFavorite{}).Where("user_id = ?", userID)
	if favoriteType != nil {
		query = query.Where("favorite_type = ?", *favoriteType)
	}
	err := query.Count(&count).Error
	return count, err
}
