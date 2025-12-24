package models

import (
	"time"

	"gorm.io/gorm"
)

type UserFollow struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	FollowerID  uint           `gorm:"not null;index" json:"follower_id"`
	FollowingID uint           `gorm:"not null;index" json:"following_id"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Follower  User `gorm:"foreignKey:FollowerID" json:"follower,omitempty"`
	Following User `gorm:"foreignKey:FollowingID" json:"following,omitempty"`
}

// TableName specifies the table name for UserFollow
func (UserFollow) TableName() string {
	return "user_follows"
}

