package models

import (
	"time"

	"gorm.io/gorm"
)

type BulkUploadEntityType string

const (
	BulkUploadEntityTypeRoute   BulkUploadEntityType = "route"
	BulkUploadEntityTypeStop    BulkUploadEntityType = "stop"
	BulkUploadEntityTypeVehicle BulkUploadEntityType = "vehicle"
)

type BulkUploadStatus string

const (
	BulkUploadStatusPending    BulkUploadStatus = "pending"
	BulkUploadStatusProcessing BulkUploadStatus = "processing"
	BulkUploadStatusCompleted  BulkUploadStatus = "completed"
	BulkUploadStatusFailed     BulkUploadStatus = "failed"
)

type BulkUploadLog struct {
	ID              uint                `gorm:"primaryKey" json:"id"`
	EntityType      BulkUploadEntityType `gorm:"type:varchar(20);not null" json:"entity_type"`
	FilePath        string              `gorm:"not null" json:"file_path"`
	Status          BulkUploadStatus    `gorm:"type:varchar(20);default:'pending'" json:"status"`
	TotalRows       int                 `json:"total_rows"`
	SuccessCount    int                 `json:"success_count"`
	DuplicateCount  int                 `json:"duplicate_count"`
	ErrorCount      int                 `json:"error_count"`
	ErrorMessage    *string             `gorm:"type:text" json:"error_message,omitempty"`
	UserID          uint                `gorm:"not null;index" json:"user_id"`
	LastProcessedRow int                `json:"last_processed_row"` // Track progress for resume
	LastUpdatedAt   time.Time           `gorm:"index" json:"last_updated_at"` // For stuck job detection
	JobID           *string             `gorm:"index" json:"job_id,omitempty"` // asynq job ID for recovery
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	DeletedAt       gorm.DeletedAt      `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

