package models

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	RoleCommonUser UserRole = "common_user"
	RoleAdmin      UserRole = "admin"
)

type OAuthProvider string

const (
	OAuthProviderGoogle OAuthProvider = "google"
	OAuthProviderGithub OAuthProvider = "github"
)

type User struct {
	ID            uint         `gorm:"primaryKey" json:"id"`
	Email         string       `gorm:"uniqueIndex;not null" json:"email"`
	Username      string       `gorm:"uniqueIndex;not null" json:"username"`
	Password      *string      `gorm:"type:varchar(255)" json:"-"` // Nullable for OAuth users
	OAuthProvider *OAuthProvider `gorm:"type:varchar(50)" json:"oauth_provider,omitempty"`
	OAuthID       *string      `gorm:"type:varchar(255)" json:"oauth_id,omitempty"`
	Role          UserRole     `gorm:"type:varchar(20);default:'common_user'" json:"role"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) IsOAuthUser() bool {
	return u.OAuthProvider != nil && u.OAuthID != nil
}

