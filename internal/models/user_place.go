package models

import (
	"time"
)

type PlaceType string

const (
	PlaceTypeHome       PlaceType = "home"
	PlaceTypeOffice     PlaceType = "office"
	PlaceTypeSchool     PlaceType = "school"
	PlaceTypeGym        PlaceType = "gym"
	PlaceTypeShopping   PlaceType = "shopping"
	PlaceTypeRestaurant PlaceType = "restaurant"
	PlaceTypeHospital   PlaceType = "hospital"
	PlaceTypeOther      PlaceType = "other"
	PlaceTypeCustom     PlaceType = "custom"
)

// UserPlace represents a user's saved place (home, office, etc.)
type UserPlace struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	PlaceType PlaceType `gorm:"type:varchar(50);not null" json:"place_type"`
	Name      string    `gorm:"not null" json:"name"`
	Latitude  float64   `gorm:"not null" json:"latitude"`
	Longitude float64   `gorm:"not null" json:"longitude"`
	Address   string    `gorm:"type:text" json:"address,omitempty"`
	Notes     string    `gorm:"type:text" json:"notes,omitempty"`
	IsDefault bool      `gorm:"default:false" json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName returns the table name for GORM
func (UserPlace) TableName() string {
	return "user_places"
}
