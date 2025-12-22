package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type RouteStopRepository interface {
	Create(routeStop *models.RouteStop) error
	FindByRouteID(routeID uint) ([]models.RouteStop, error)
	FindByID(id uint) (*models.RouteStop, error)
	Update(routeStop *models.RouteStop) error
	Delete(id uint) error
	DeleteByRouteID(routeID uint) error
	DeleteByRouteAndStop(routeID, stopID uint) error
}

type routeStopRepository struct {
	db *gorm.DB
}

func NewRouteStopRepository(db *gorm.DB) RouteStopRepository {
	return &routeStopRepository{db: db}
}

func (r *routeStopRepository) Create(routeStop *models.RouteStop) error {
	return r.db.Create(routeStop).Error
}

func (r *routeStopRepository) FindByRouteID(routeID uint) ([]models.RouteStop, error) {
	var routeStops []models.RouteStop
	err := r.db.Where("route_id = ?", routeID).
		Order("sequence_order ASC").
		Preload("Stop").
		Find(&routeStops).Error
	return routeStops, err
}

func (r *routeStopRepository) FindByID(id uint) (*models.RouteStop, error) {
	var routeStop models.RouteStop
	err := r.db.First(&routeStop, id).Error
	if err != nil {
		return nil, err
	}
	return &routeStop, nil
}

func (r *routeStopRepository) Update(routeStop *models.RouteStop) error {
	return r.db.Save(routeStop).Error
}

func (r *routeStopRepository) Delete(id uint) error {
	return r.db.Delete(&models.RouteStop{}, id).Error
}

func (r *routeStopRepository) DeleteByRouteID(routeID uint) error {
	return r.db.Where("route_id = ?", routeID).Delete(&models.RouteStop{}).Error
}

func (r *routeStopRepository) DeleteByRouteAndStop(routeID, stopID uint) error {
	return r.db.Where("route_id = ? AND stop_id = ?", routeID, stopID).
		Delete(&models.RouteStop{}).Error
}

