package models

import (
	"time"

	"gorm.io/gorm"
)

type StopType string

const (
	StopTypeStop     StopType = "stop"
	StopTypeTerminal StopType = "terminal"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

type Stop struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Name       string         `gorm:"not null" json:"name"`
	Type       StopType       `gorm:"type:varchar(20);not null" json:"type"`
	Latitude   float64        `gorm:"not null" json:"latitude"`
	Longitude  float64        `gorm:"not null" json:"longitude"`
	Address    string         `json:"address"`
	City       string         `json:"city"`
	District   string         `json:"district"`
	Facilities *string        `gorm:"type:jsonb" json:"facilities,omitempty"` // JSON string
	Status     Status         `gorm:"type:varchar(20);default:'active'" json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	RouteStops []RouteStop `gorm:"foreignKey:StopID" json:"route_stops,omitempty"`
}
