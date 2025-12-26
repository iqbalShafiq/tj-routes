package models

import (
	"time"

	"gorm.io/gorm"
)

type PostType string

const (
	PostTypeDiscussion   PostType = "discussion"
	PostTypeInfo         PostType = "info"
	PostTypeQuestion     PostType = "question"
	PostTypeAnnouncement PostType = "announcement"
)

type ForumPost struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	ForumID       uint           `gorm:"not null;index" json:"forum_id"`
	UserID        uint           `gorm:"not null;index" json:"user_id"`
	PostType      PostType       `gorm:"type:varchar(50);not null;default:'discussion'" json:"post_type"`
	Title         string         `gorm:"not null" json:"title"`
	Content       string         `gorm:"type:text;not null" json:"content"`
	LinkedReportID *uint         `gorm:"index" json:"linked_report_id,omitempty"`
	IsPinned      bool           `gorm:"default:false" json:"is_pinned"`
	PhotoURLs     *string        `gorm:"type:jsonb" json:"photo_urls,omitempty"` // JSON array of URLs
	PDFURLs       *string         `gorm:"type:jsonb" json:"pdf_urls,omitempty"`   // JSON array of URLs
	Upvotes       int            `gorm:"default:0" json:"upvotes"`
	Downvotes     int            `gorm:"default:0" json:"downvotes"`
	CommentCount  int            `gorm:"default:0" json:"comment_count"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Forum       Forum         `gorm:"foreignKey:ForumID" json:"forum,omitempty"`
	User        User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	LinkedReport *Report      `gorm:"foreignKey:LinkedReportID" json:"linked_report,omitempty"`
	Hashtags    []ForumPostHashtag `gorm:"foreignKey:ForumPostID" json:"hashtags,omitempty"`
}

type ForumPostHashtag struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ForumPostID uint      `gorm:"not null;index" json:"forum_post_id"`
	HashtagID  uint      `gorm:"not null;index" json:"hashtag_id"`
	CreatedAt  time.Time `json:"created_at"`

	// Relationships
	ForumPost ForumPost `gorm:"foreignKey:ForumPostID" json:"-"`
	Hashtag Hashtag `gorm:"foreignKey:HashtagID" json:"hashtag,omitempty"`
}

// TableName specifies the table name for ForumPostHashtag
func (ForumPostHashtag) TableName() string {
	return "forum_post_hashtags"
}

