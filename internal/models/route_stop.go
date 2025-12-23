package models

import (
	"time"
)

type RouteStop struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	RouteID        uint      `gorm:"not null;index" json:"route_id"`
	StopID         uint      `gorm:"not null;index" json:"stop_id"`
	SequenceOrder  int       `gorm:"not null" json:"sequence_order"`
	IsOrigin       bool      `gorm:"default:false" json:"is_origin"`
	IsDestination  bool      `gorm:"default:false" json:"is_destination"`
	CreatedAt      time.Time `json:"created_at"`

	// Relationships
	Stop Stop `gorm:"foreignKey:StopID" json:"stop,omitempty"`
}

// Ensure unique combination of route and sequence order
func (RouteStop) TableName() string {
	return "route_stops"
}

