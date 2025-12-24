package models

import (
	"time"

	"gorm.io/gorm"
)

type Hashtag struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Name       string         `gorm:"uniqueIndex;not null;type:varchar(100)" json:"name"`
	UsageCount int            `gorm:"default:0" json:"usage_count"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Reports []ReportHashtag `gorm:"foreignKey:HashtagID" json:"-"`
}

type ReportHashtag struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ReportID   uint      `gorm:"not null;index" json:"report_id"`
	HashtagID  uint      `gorm:"not null;index" json:"hashtag_id"`
	CreatedAt  time.Time `json:"created_at"`

	// Relationships
	Report  Report  `gorm:"foreignKey:ReportID" json:"-"`
	Hashtag Hashtag `gorm:"foreignKey:HashtagID" json:"hashtag,omitempty"`
}

// TableName specifies the table name for ReportHashtag
func (ReportHashtag) TableName() string {
	return "report_hashtags"
}

