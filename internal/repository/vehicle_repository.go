package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type VehicleRepository interface {
	Create(vehicle *models.Vehicle) error
	FindByID(id uint) (*models.Vehicle, error)
	FindByVehiclePlate(plate string) (*models.Vehicle, error)
	Update(vehicle *models.Vehicle) error
	Delete(id uint) error
	List(offset, limit int, filters map[string]interface{}) ([]models.Vehicle, int64, error)
}

type vehicleRepository struct {
	db *gorm.DB
}

func NewVehicleRepository(db *gorm.DB) VehicleRepository {
	return &vehicleRepository{db: db}
}

func (r *vehicleRepository) Create(vehicle *models.Vehicle) error {
	return r.db.Create(vehicle).Error
}

func (r *vehicleRepository) FindByID(id uint) (*models.Vehicle, error) {
	var vehicle models.Vehicle
	err := r.db.Preload("Route").First(&vehicle, id).Error
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

func (r *vehicleRepository) FindByVehiclePlate(plate string) (*models.Vehicle, error) {
	var vehicle models.Vehicle
	err := r.db.Where("LOWER(vehicle_plate) = LOWER(?)", plate).First(&vehicle).Error
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

func (r *vehicleRepository) Update(vehicle *models.Vehicle) error {
	return r.db.Save(vehicle).Error
}

func (r *vehicleRepository) Delete(id uint) error {
	return r.db.Delete(&models.Vehicle{}, id).Error
}

func (r *vehicleRepository) List(offset, limit int, filters map[string]interface{}) ([]models.Vehicle, int64, error) {
	var vehicles []models.Vehicle
	var total int64

	query := r.db.Model(&models.Vehicle{})

	if status, ok := filters["status"].(models.Status); ok {
		query = query.Where("status = ?", status)
	}
	if routeID, ok := filters["route_id"].(uint); ok && routeID > 0 {
		query = query.Where("route_id = ?", routeID)
	}

	// Fuzzy search using pg_trgm with similarity threshold and ILIKE for partial matching
	if search, ok := filters["search"].(string); ok && search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"(similarity(vehicle_plate, ?) > 0.2 OR similarity(vehicle_type, ?) > 0.2) OR (vehicle_plate ILIKE ? OR vehicle_type ILIKE ?)",
			search, search, searchPattern, searchPattern,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Note: We skip relevance-based ordering for now due to GORM limitations with parameterized ORDER BY
	// The search filtering still works correctly via the WHERE clause

	err := query.Preload("Route").Offset(offset).Limit(limit).Find(&vehicles).Error
	return vehicles, total, err
}

