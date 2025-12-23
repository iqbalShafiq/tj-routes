package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type StopRepository interface {
	Create(stop *models.Stop) error
	FindByID(id uint) (*models.Stop, error)
	FindByLatitudeAndLongitude(lat, lng float64) (*models.Stop, error)
	Update(stop *models.Stop) error
	Delete(id uint) error
	List(offset, limit int, filters map[string]interface{}) ([]models.Stop, int64, error)
}

type stopRepository struct {
	db *gorm.DB
}

func NewStopRepository(db *gorm.DB) StopRepository {
	return &stopRepository{db: db}
}

func (r *stopRepository) Create(stop *models.Stop) error {
	return r.db.Create(stop).Error
}

func (r *stopRepository) FindByID(id uint) (*models.Stop, error) {
	var stop models.Stop
	err := r.db.First(&stop, id).Error
	if err != nil {
		return nil, err
	}
	return &stop, nil
}

func (r *stopRepository) FindByLatitudeAndLongitude(lat, lng float64) (*models.Stop, error) {
	var stop models.Stop
	// Tolerance: 0.0001 degrees ≈ 11 meters
	tolerance := 0.0001
	err := r.db.Where(
		"ABS(latitude - ?) < ? AND ABS(longitude - ?) < ?",
		lat, tolerance, lng, tolerance,
	).First(&stop).Error
	if err != nil {
		return nil, err
	}
	return &stop, nil
}

func (r *stopRepository) Update(stop *models.Stop) error {
	return r.db.Save(stop).Error
}

func (r *stopRepository) Delete(id uint) error {
	return r.db.Delete(&models.Stop{}, id).Error
}

func (r *stopRepository) List(offset, limit int, filters map[string]interface{}) ([]models.Stop, int64, error) {
	var stops []models.Stop
	var total int64

	query := r.db.Model(&models.Stop{})

	// Apply filters
	if status, ok := filters["status"].(models.Status); ok {
		query = query.Where("status = ?", status)
	}
	if stopType, ok := filters["type"].(models.StopType); ok {
		query = query.Where("type = ?", stopType)
	}
	if city, ok := filters["city"].(string); ok && city != "" {
		query = query.Where("city = ?", city)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Find(&stops).Error
	return stops, total, err
}

