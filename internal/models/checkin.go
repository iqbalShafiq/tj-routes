package models

import (
	"time"

	"gorm.io/gorm"
)

type CheckInStatus string

const (
	CheckInStatusInProgress CheckInStatus = "in_progress"
	CheckInStatusCompleted  CheckInStatus = "completed"
	CheckInStatusCancelled  CheckInStatus = "cancelled"
)

type CheckIn struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"not null;index" json:"user_id"`
	RouteID      uint           `gorm:"not null;index" json:"route_id"`
	StartStopID  uint           `gorm:"not null;index" json:"start_stop_id"`
	EndStopID    *uint          `gorm:"index" json:"end_stop_id,omitempty"`
	StartTime    time.Time      `gorm:"not null" json:"start_time"`
	EndTime      *time.Time     `json:"end_time,omitempty"`
	Duration     *int           `gorm:"type:int" json:"duration_seconds,omitempty"`
	Notes        *string        `gorm:"type:text" json:"notes,omitempty"`
	Status       CheckInStatus  `gorm:"type:varchar(20);default:'in_progress'" json:"status"`
	PointsEarned int            `gorm:"default:0" json:"points_earned"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User      User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Route     Route    `gorm:"foreignKey:RouteID" json:"route,omitempty"`
	StartStop Stop     `gorm:"foreignKey:StartStopID" json:"start_stop,omitempty"`
	EndStop   *Stop    `gorm:"foreignKey:EndStopID" json:"end_stop,omitempty"`
}

// TableName specifies the table name for GORM
func (CheckIn) TableName() string {
	return "check_ins"
}

// BeforeCreate sets default values before creating a check-in
func (c *CheckIn) BeforeCreate(tx *gorm.DB) error {
	if c.Status == "" {
		c.Status = CheckInStatusInProgress
	}
	// Ensure EndStopID is nil when not set to avoid FK violation
	if c.EndStopID != nil && *c.EndStopID == 0 {
		c.EndStopID = nil
	}
	return nil
}
