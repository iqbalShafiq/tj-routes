package models

import (
	"time"

	"gorm.io/gorm"
)

type MessageReaction struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	MessageID uint           `gorm:"not null;uniqueIndex:idx_message_user" json:"message_id"`
	UserID    uint           `gorm:"not null;uniqueIndex:idx_message_user" json:"user_id"`
	Emoji     string         `gorm:"type:varchar(50);not null" json:"emoji"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Message Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (MessageReaction) TableName() string {
	return "message_reactions"
}
