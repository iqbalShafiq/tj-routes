package models

import (
	"time"

	"gorm.io/gorm"
)

type ForumMessage struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ForumID   uint           `gorm:"not null;index" json:"forum_id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Forum Forum `gorm:"foreignKey:ForumID" json:"forum,omitempty"`
	User  User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ForumMessage) TableName() string {
	return "forum_messages"
}
