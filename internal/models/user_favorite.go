package models

import (
	"time"
)

type FavoriteType string

const (
	FavoriteTypeRoute FavoriteType = "route"
	FavoriteTypeStop  FavoriteType = "stop"
)

// UserFavorite represents a user's favorite route or stop
type UserFavorite struct {
	ID           uint          `gorm:"primaryKey" json:"id"`
	UserID       uint          `gorm:"not null;index" json:"user_id"`
	FavoriteType FavoriteType  `gorm:"type:varchar(20);not null" json:"favorite_type"`
	RouteID      *uint         `gorm:"index" json:"route_id,omitempty"`
	StopID       *uint         `gorm:"index" json:"stop_id,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`

	// Relationships
	User  User   `gorm:"foreignKey:UserID" json:"-"`
	Route *Route `gorm:"foreignKey:RouteID" json:"route,omitempty"`
	Stop  *Stop  `gorm:"foreignKey:StopID" json:"stop,omitempty"`
}

// TableName returns the table name for GORM
func (UserFavorite) TableName() string {
	return "user_favorites"
}
