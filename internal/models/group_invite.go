package models

import (
	"time"

	"gorm.io/gorm"
)

type GroupInvite struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	GroupID   uint           `gorm:"not null;index" json:"group_id"`
	InviterID uint           `gorm:"not null" json:"inviter_id"`
	InviteeID uint           `gorm:"not null;index" json:"invitee_id"`
	Status    string         `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Group   GroupChat `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	Inviter User      `gorm:"foreignKey:InviterID" json:"inviter,omitempty"`
	Invitee User      `gorm:"foreignKey:InviteeID" json:"invitee,omitempty"`
}

func (GroupInvite) TableName() string {
	return "group_invites"
}
