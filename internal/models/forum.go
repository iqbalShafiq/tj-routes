package models

import (
	"time"

	"gorm.io/gorm"
)

type Forum struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	RouteID uint           `gorm:"not null;uniqueIndex;index" json:"route_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Route  Route         `gorm:"foreignKey:RouteID" json:"route,omitempty"`
	Posts  []ForumPost   `gorm:"foreignKey:ForumID" json:"posts,omitempty"`
	Members []ForumMember `gorm:"foreignKey:ForumID" json:"members,omitempty"`
}

