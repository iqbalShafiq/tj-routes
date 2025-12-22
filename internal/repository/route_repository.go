package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type RouteRepository interface {
	Create(route *models.Route) error
	FindByID(id uint) (*models.Route, error)
	Update(route *models.Route) error
	Delete(id uint) error
	List(offset, limit int, filters map[string]interface{}) ([]models.Route, int64, error)
}

type routeRepository struct {
	db *gorm.DB
}

func NewRouteRepository(db *gorm.DB) RouteRepository {
	return &routeRepository{db: db}
}

func (r *routeRepository) Create(route *models.Route) error {
	return r.db.Create(route).Error
}

func (r *routeRepository) FindByID(id uint) (*models.Route, error) {
	var route models.Route
	err := r.db.Preload("RouteStops.Stop").First(&route, id).Error
	if err != nil {
		return nil, err
	}
	return &route, nil
}

func (r *routeRepository) Update(route *models.Route) error {
	return r.db.Save(route).Error
}

func (r *routeRepository) Delete(id uint) error {
	return r.db.Delete(&models.Route{}, id).Error
}

func (r *routeRepository) List(offset, limit int, filters map[string]interface{}) ([]models.Route, int64, error) {
	var routes []models.Route
	var total int64

	query := r.db.Model(&models.Route{})

	if status, ok := filters["status"].(models.Status); ok {
		query = query.Where("status = ?", status)
	}
	if routeNumber, ok := filters["route_number"].(string); ok && routeNumber != "" {
		query = query.Where("route_number = ?", routeNumber)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("RouteStops.Stop").Offset(offset).Limit(limit).Find(&routes).Error
	return routes, total, err
}

