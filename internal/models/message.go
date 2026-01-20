package models

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ConversationID *uint          `gorm:"index" json:"conversation_id,omitempty"`
	GroupID        *uint          `gorm:"index" json:"group_id,omitempty"`
	SenderID       uint           `gorm:"not null;index" json:"sender_id"`
	Content        string         `gorm:"type:text;not null" json:"content"`
	MessageType    string         `gorm:"type:varchar(20);not null;default:'text'" json:"message_type"`
	ReplyToID      *uint          `gorm:"index" json:"reply_to_id,omitempty"`
	Status         string         `gorm:"type:varchar(20);not null;default:'sent'" json:"status"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Conversation *Conversation     `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`
	Group        *GroupChat        `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	Sender       User              `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	ReplyTo      *Message          `gorm:"foreignKey:ReplyToID" json:"reply_to,omitempty"`
	Reactions    []MessageReaction `gorm:"foreignKey:MessageID" json:"reactions,omitempty"`
}
