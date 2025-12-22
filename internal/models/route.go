package models

import (
	"time"

	"gorm.io/gorm"
)

type Route struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	RouteNumber string         `gorm:"uniqueIndex;not null" json:"route_number"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	Status      Status         `gorm:"type:varchar(20);default:'active'" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	RouteStops []RouteStop `gorm:"foreignKey:RouteID;order:sequence_order" json:"route_stops,omitempty"`
	Vehicles   []Vehicle   `gorm:"foreignKey:RouteID" json:"vehicles,omitempty"`
}
