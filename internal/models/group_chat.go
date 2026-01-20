package models

import (
	"time"

	"gorm.io/gorm"
)

type GroupChat struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	AvatarURL   *string        `json:"avatar_url"`
	Type        string         `gorm:"type:varchar(20);not null;default:'private'" json:"type"`
	MaxMembers  int            `gorm:"default:500" json:"max_members"`
	CreatedByID uint           `gorm:"not null" json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	CreatedBy User          `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	Members   []GroupMember `gorm:"foreignKey:GroupID" json:"members,omitempty"`
	Messages  []Message     `gorm:"foreignKey:GroupID" json:"messages,omitempty"`
	Invites   []GroupInvite `gorm:"foreignKey:GroupID" json:"invites,omitempty"`
}
