package models

import (
	"time"
)

type BadgeCriteriaType string

const (
	BadgeCriteriaReportsAccepted   BadgeCriteriaType = "reports_accepted"
	BadgeCriteriaCommentsMade      BadgeCriteriaType = "comments_made"
	BadgeCriteriaUpvotesReceived   BadgeCriteriaType = "upvotes_received"
	BadgeCriteriaReputationPoints  BadgeCriteriaType = "reputation_points"
	BadgeCriteriaCheckInsCount     BadgeCriteriaType = "check_ins_count"
	BadgeCriteriaUniqueRoutes      BadgeCriteriaType = "unique_routes"
	BadgeCriteriaConsecutiveDays   BadgeCriteriaType = "consecutive_days"
)

type Badge struct {
	ID            uint              `gorm:"primaryKey" json:"id"`
	Name          string            `gorm:"not null" json:"name"`
	Description   string            `gorm:"type:text" json:"description"`
	Icon          string            `gorm:"type:varchar(255)" json:"icon"` // URL or emoji
	CriteriaType  BadgeCriteriaType `gorm:"type:varchar(50);not null" json:"criteria_type"`
	CriteriaValue int               `gorm:"not null" json:"criteria_value"` // Threshold to earn
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type UserBadge struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	UserID   uint      `gorm:"not null;index" json:"user_id"`
	BadgeID  uint      `gorm:"not null;index" json:"badge_id"`
	EarnedAt time.Time `gorm:"not null" json:"earned_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Badge Badge `gorm:"foreignKey:BadgeID" json:"badge,omitempty"`
}

