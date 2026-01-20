package models

import (
	"time"

	"gorm.io/gorm"
)

type GroupMember struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	GroupID    uint           `gorm:"not null;uniqueIndex:idx_group_user" json:"group_id"`
	UserID     uint           `gorm:"not null;uniqueIndex:idx_group_user" json:"user_id"`
	Role       string         `gorm:"type:varchar(20);not null;default:'member'" json:"role"`
	LastReadAt *time.Time     `json:"last_read_at"`
	MutedUntil *time.Time     `json:"muted_until"`
	JoinedAt   time.Time      `gorm:"not null" json:"joined_at"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	Group GroupChat `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	User  User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (GroupMember) TableName() string {
	return "group_members"
}
