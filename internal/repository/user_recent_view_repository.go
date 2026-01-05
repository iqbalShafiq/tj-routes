package repository

import (
	"time"

	"tj-routes/internal/models"
	"gorm.io/gorm"
)

// UserRecentViewRepository interface for user recent view operations
type UserRecentViewRepository interface {
	Create(view *models.UserRecentView) error
	Upsert(view *models.UserRecentView) error
	FindByUserID(userID uint, viewType models.RecentViewType, offset, limit int) ([]models.UserRecentView, int64, error)
	DeleteOldViews(userID uint, keepCount int) error
	CleanupOldViews(olderThan time.Duration) error
	CountByUserID(userID uint, viewType *models.RecentViewType) (int64, error)
}

type userRecentViewRepository struct {
	db *gorm.DB
}

// NewUserRecentViewRepository creates a new UserRecentViewRepository
func NewUserRecentViewRepository(db *gorm.DB) UserRecentViewRepository {
	return &userRecentViewRepository{db: db}
}

// Create creates a new user recent view
func (r *userRecentViewRepository) Create(view *models.UserRecentView) error {
	return r.db.Create(view).Error
}

// Upsert creates or updates a user recent view (based on unique constraint)
func (r *userRecentViewRepository) Upsert(view *models.UserRecentView) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Try to find existing record
		var existing models.UserRecentView
		query := tx.Where("user_id = ? AND view_type = ?", view.UserID, view.ViewType)

		if view.RouteID != nil {
			query = query.Where("route_id = ?", *view.RouteID)
		} else {
			query = query.Where("route_id IS NULL")
		}

		if view.FromStopID != nil {
			query = query.Where("from_stop_id = ?", *view.FromStopID)
		} else {
			query = query.Where("from_stop_id IS NULL")
		}

		if view.ToStopID != nil {
			query = query.Where("to_stop_id = ?", *view.ToStopID)
		} else {
			query = query.Where("to_stop_id IS NULL")
		}

		err := query.First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			// Create new record
			return tx.Create(view).Error
		} else if err != nil {
			return err
		}

		// Update existing record's viewed_at timestamp
		return tx.Model(&existing).Update("viewed_at", view.ViewedAt).Error
	})
}

// FindByUserID finds all recent views by user ID with pagination
func (r *userRecentViewRepository) FindByUserID(userID uint, viewType models.RecentViewType, offset, limit int) ([]models.UserRecentView, int64, error) {
	var views []models.UserRecentView
	var total int64

	query := r.db.Model(&models.UserRecentView{}).Where("user_id = ? AND view_type = ?", userID, viewType)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with preloads
	err := query.Offset(offset).Limit(limit).Order("viewed_at DESC").
		Preload("Route").
		Preload("FromStop").
		Preload("ToStop").
		Find(&views).Error

	return views, total, err
}

// DeleteOldViews deletes old views keeping only the most recent ones
func (r *userRecentViewRepository) DeleteOldViews(userID uint, keepCount int) error {
	// Get the IDs to keep
	var keepIDs []uint
	subQuery := r.db.Model(&models.UserRecentView{}).
		Where("user_id = ?", userID).
		Order("viewed_at DESC").
		Limit(keepCount).
		Select("id")

	if err := r.db.Model(&models.UserRecentView{}).
		Where("id IN ?", subQuery).
		Pluck("id", &keepIDs).Error; err != nil {
		return err
	}

	if len(keepIDs) == 0 {
		return nil
	}

	// Delete all other views
	return r.db.Where("user_id = ? AND id NOT IN ?", userID, keepIDs).
		Delete(&models.UserRecentView{}).Error
}

// CleanupOldViews deletes views older than the specified duration
func (r *userRecentViewRepository) CleanupOldViews(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	return r.db.Where("viewed_at < ?", cutoff).
		Delete(&models.UserRecentView{}).Error
}

// CountByUserID counts recent views by user ID
func (r *userRecentViewRepository) CountByUserID(userID uint, viewType *models.RecentViewType) (int64, error) {
	var count int64
	query := r.db.Model(&models.UserRecentView{}).Where("user_id = ?", userID)
	if viewType != nil {
		query = query.Where("view_type = ?", *viewType)
	}
	err := query.Count(&count).Error
	return count, err
}
