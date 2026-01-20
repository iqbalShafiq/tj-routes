package models

import (
	"time"

	"gorm.io/gorm"
)

type ChatRequest struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	SenderID   uint           `gorm:"not null;index" json:"sender_id"`
	ReceiverID uint           `gorm:"not null;index" json:"receiver_id"`
	Message    string         `gorm:"type:text;not null" json:"message"`
	Status     string         `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	Sender   User `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	Receiver User `gorm:"foreignKey:ReceiverID" json:"receiver,omitempty"`
}

func (ChatRequest) TableName() string {
	return "chat_requests"
}
