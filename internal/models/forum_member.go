package models

import (
	"time"

	"gorm.io/gorm"
)

type ForumMember struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ForumID   uint           `gorm:"not null;index" json:"forum_id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Forum Forum `gorm:"foreignKey:ForumID" json:"forum,omitempty"`
	User  User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name for ForumMember
func (ForumMember) TableName() string {
	return "forum_members"
}

