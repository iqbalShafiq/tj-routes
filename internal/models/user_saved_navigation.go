package models

import (
	"time"
)

// UserSavedNavigation represents a user's saved navigation pair (e.g., "home to office")
type UserSavedNavigation struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	Name          string    `gorm:"type:varchar(255)" json:"name,omitempty"`
	FromPlaceType *string   `gorm:"type:varchar(50)" json:"from_place_type,omitempty"`
	FromPlaceID   *uint     `gorm:"index" json:"from_place_id,omitempty"`
	FromStopID    *uint     `gorm:"index" json:"from_stop_id,omitempty"`
	ToPlaceType   *string   `gorm:"type:varchar(50)" json:"to_place_type,omitempty"`
	ToPlaceID     *uint     `gorm:"index" json:"to_place_id,omitempty"`
	ToStopID      *uint     `gorm:"index" json:"to_stop_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships
	User      User       `gorm:"foreignKey:UserID" json:"-"`
	FromPlace *UserPlace `gorm:"foreignKey:FromPlaceID" json:"from_place,omitempty"`
	FromStop  *Stop      `gorm:"foreignKey:FromStopID" json:"from_stop,omitempty"`
	ToPlace   *UserPlace `gorm:"foreignKey:ToPlaceID" json:"to_place,omitempty"`
	ToStop    *Stop      `gorm:"foreignKey:ToStopID" json:"to_stop,omitempty"`
}

// TableName returns the table name for GORM
func (UserSavedNavigation) TableName() string {
	return "user_saved_navigations"
}
