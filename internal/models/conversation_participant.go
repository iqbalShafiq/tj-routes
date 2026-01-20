package models

import (
	"time"

	"gorm.io/gorm"
)

type ConversationParticipant struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ConversationID uint           `gorm:"not null;uniqueIndex:idx_conversation_user" json:"conversation_id"`
	UserID         uint           `gorm:"not null;uniqueIndex:idx_conversation_user" json:"user_id"`
	LastReadAt     *time.Time     `json:"last_read_at"`
	JoinedAt       time.Time      `gorm:"not null" json:"joined_at"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Conversation Conversation `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`
	User         User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ConversationParticipant) TableName() string {
	return "conversation_participants"
}
