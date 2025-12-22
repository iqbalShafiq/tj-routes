package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type RouteChangeRepository interface {
	Create(routeChange *models.RouteChange) error
	FindByRouteID(routeID uint) ([]models.RouteChange, error)
	FindByID(id uint) (*models.RouteChange, error)
}

type routeChangeRepository struct {
	db *gorm.DB
}

func NewRouteChangeRepository(db *gorm.DB) RouteChangeRepository {
	return &routeChangeRepository{db: db}
}

func (r *routeChangeRepository) Create(routeChange *models.RouteChange) error {
	return r.db.Create(routeChange).Error
}

func (r *routeChangeRepository) FindByRouteID(routeID uint) ([]models.RouteChange, error) {
	var routeChanges []models.RouteChange
	err := r.db.Where("route_id = ?", routeID).
		Preload("ChangedByUser").
		Preload("AffectedStop").
		Order("created_at DESC").
		Find(&routeChanges).Error
	return routeChanges, err
}

func (r *routeChangeRepository) FindByID(id uint) (*models.RouteChange, error) {
	var routeChange models.RouteChange
	err := r.db.Preload("ChangedByUser").
		Preload("AffectedStop").
		First(&routeChange, id).Error
	if err != nil {
		return nil, err
	}
	return &routeChange, nil
}

