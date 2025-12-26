package models

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ReportID    *uint          `gorm:"index" json:"report_id,omitempty"` // Nullable for forum post comments
	ForumPostID *uint          `gorm:"index" json:"forum_post_id,omitempty"` // Nullable for report comments
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	ParentID    *uint          `gorm:"index" json:"parent_id,omitempty"` // nil = top-level comment
	Content     string         `gorm:"type:text;not null" json:"content"`
	Upvotes     int            `gorm:"default:0" json:"upvotes"`
	Downvotes   int            `gorm:"default:0" json:"downvotes"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Report    *Report   `gorm:"foreignKey:ReportID" json:"report,omitempty"`
	ForumPost *ForumPost `gorm:"foreignKey:ForumPostID" json:"forum_post,omitempty"`
	Replies   []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

