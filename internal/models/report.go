package models

import (
	"time"

	"gorm.io/gorm"
)

type ReportType string

const (
	ReportTypeRouteIssue     ReportType = "route_issue"
	ReportTypeStopIssue      ReportType = "stop_issue"
	ReportTypeTemporaryEvent ReportType = "temporary_event"
	ReportTypePolicyChange   ReportType = "policy_change"
)

type ReportStatus string

const (
	ReportStatusPending  ReportStatus = "pending"
	ReportStatusReviewed ReportStatus = "reviewed"
	ReportStatusResolved ReportStatus = "resolved"
)

type Report struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UserID         uint           `gorm:"not null;index" json:"user_id"`
	Type           ReportType     `gorm:"type:varchar(50);not null" json:"type"`
	Title          string         `gorm:"not null" json:"title"`
	Description    string         `gorm:"type:text;not null" json:"description"`
	RelatedRouteID *uint          `gorm:"index" json:"related_route_id,omitempty"`
	RelatedStopID  *uint          `gorm:"index" json:"related_stop_id,omitempty"`
	Status         ReportStatus   `gorm:"type:varchar(20);default:'pending'" json:"status"`
	AdminNotes     *string        `gorm:"type:text" json:"admin_notes,omitempty"`
	PhotoURLs      *string        `gorm:"type:jsonb" json:"photo_urls,omitempty"` // JSON array of URLs
	PDFURLs        *string        `gorm:"type:jsonb" json:"pdf_urls,omitempty"`   // JSON array of URLs
	Upvotes        int            `gorm:"default:0" json:"upvotes"`
	Downvotes      int            `gorm:"default:0" json:"downvotes"`
	CommentCount   int            `gorm:"default:0" json:"comment_count"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User         User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	RelatedRoute *Route `gorm:"foreignKey:RelatedRouteID" json:"related_route,omitempty"`
	RelatedStop  *Stop  `gorm:"foreignKey:RelatedStopID" json:"related_stop,omitempty"`
}
