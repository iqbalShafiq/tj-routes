package models

import (
	"time"
)

type ChangeType string

const (
	ChangeTypeRouteCreated     ChangeType = "route_created"
	ChangeTypeRouteUpdated     ChangeType = "route_updated"
	ChangeTypeStopAdded        ChangeType = "stop_added"
	ChangeTypeStopRemoved      ChangeType = "stop_removed"
	ChangeTypeStopOrderChanged ChangeType = "stop_order_changed"
	ChangeTypeStopUpdated      ChangeType = "stop_updated"
)

type RouteChange struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	RouteID        uint       `gorm:"not null;index" json:"route_id"`
	ChangedBy      uint       `gorm:"not null;index" json:"changed_by"`
	ChangeType     ChangeType `gorm:"type:varchar(50);not null" json:"change_type"`
	AffectedStopID *uint      `gorm:"index" json:"affected_stop_id,omitempty"`
	OldData        *string    `gorm:"type:jsonb" json:"old_data,omitempty"` // JSON string
	NewData        *string    `gorm:"type:jsonb" json:"new_data,omitempty"` // JSON string
	Reason         string     `gorm:"type:text" json:"reason"`
	CreatedAt      time.Time  `json:"created_at"`

	// Relationships
	Route         Route `gorm:"foreignKey:RouteID" json:"route,omitempty"`
	ChangedByUser User  `gorm:"foreignKey:ChangedBy" json:"changed_by_user,omitempty"`
	AffectedStop  *Stop `gorm:"foreignKey:AffectedStopID" json:"affected_stop,omitempty"`
}
