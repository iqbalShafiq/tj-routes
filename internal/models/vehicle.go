package models

import (
	"time"

	"gorm.io/gorm"
)

type Vehicle struct {
	ID           uint         `gorm:"primaryKey" json:"id"`
	VehiclePlate string       `gorm:"uniqueIndex;not null" json:"vehicle_plate"`
	RouteID      uint         `gorm:"not null;index" json:"route_id"`
	VehicleType  string       `json:"vehicle_type"`
	Capacity     int          `json:"capacity"`
	Status       Status       `gorm:"type:varchar(20);default:'active'" json:"status"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Route Route `gorm:"foreignKey:RouteID" json:"route,omitempty"`
}

