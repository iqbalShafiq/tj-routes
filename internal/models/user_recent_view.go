package models

import (
	"time"
)

type RecentViewType string

const (
	RecentViewTypeRoute      RecentViewType = "route"
	RecentViewTypeStop       RecentViewType = "stop"
	RecentViewTypeNavigation RecentViewType = "navigation"
)

// UserRecentView represents a user's recently viewed route, stop, or navigation
type UserRecentView struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uint           `gorm:"not null;index" json:"user_id"`
	ViewType   RecentViewType `gorm:"type:varchar(20);not null" json:"view_type"`
	RouteID    *uint          `gorm:"index" json:"route_id,omitempty"`
	FromStopID *uint          `gorm:"index" json:"from_stop_id,omitempty"`
	ToStopID   *uint          `gorm:"index" json:"to_stop_id,omitempty"`
	ViewedAt   time.Time      `gorm:"not null;index" json:"viewed_at"`

	// Relationships
	User     User   `gorm:"foreignKey:UserID" json:"-"`
	Route    *Route `gorm:"foreignKey:RouteID" json:"route,omitempty"`
	FromStop *Stop  `gorm:"foreignKey:FromStopID" json:"from_stop,omitempty"`
	ToStop   *Stop  `gorm:"foreignKey:ToStopID" json:"to_stop,omitempty"`
}

// TableName returns the table name for GORM
func (UserRecentView) TableName() string {
	return "user_recent_views"
}
