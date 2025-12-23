package repository

import (
	"time"

	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type BulkUploadLogRepository interface {
	Create(log *models.BulkUploadLog) error
	FindByID(id uint) (*models.BulkUploadLog, error)
	Update(log *models.BulkUploadLog) error
	Delete(id uint) error
	List(offset, limit int, filters map[string]interface{}) ([]models.BulkUploadLog, int64, error)
	FindStuckJobs(thresholdMinutes int) ([]models.BulkUploadLog, error)
	FindByJobID(jobID string) (*models.BulkUploadLog, error)
}

type bulkUploadLogRepository struct {
	db *gorm.DB
}

func NewBulkUploadLogRepository(db *gorm.DB) BulkUploadLogRepository {
	return &bulkUploadLogRepository{db: db}
}

func (r *bulkUploadLogRepository) Create(log *models.BulkUploadLog) error {
	return r.db.Create(log).Error
}

func (r *bulkUploadLogRepository) FindByID(id uint) (*models.BulkUploadLog, error) {
	var log models.BulkUploadLog
	err := r.db.Preload("User").First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *bulkUploadLogRepository) Update(log *models.BulkUploadLog) error {
	return r.db.Save(log).Error
}

func (r *bulkUploadLogRepository) Delete(id uint) error {
	return r.db.Delete(&models.BulkUploadLog{}, id).Error
}

func (r *bulkUploadLogRepository) List(offset, limit int, filters map[string]interface{}) ([]models.BulkUploadLog, int64, error) {
	var logs []models.BulkUploadLog
	var total int64

	query := r.db.Model(&models.BulkUploadLog{})

	// Apply filters
	if entityType, ok := filters["entity_type"].(models.BulkUploadEntityType); ok {
		query = query.Where("entity_type = ?", entityType)
	}
	if status, ok := filters["status"].(models.BulkUploadStatus); ok {
		query = query.Where("status = ?", status)
	}
	if userID, ok := filters["user_id"].(uint); ok && userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}

func (r *bulkUploadLogRepository) FindStuckJobs(thresholdMinutes int) ([]models.BulkUploadLog, error) {
	var logs []models.BulkUploadLog
	threshold := time.Now().Add(-time.Duration(thresholdMinutes) * time.Minute)
	
	err := r.db.Where("status = ? AND last_updated_at < ?", models.BulkUploadStatusProcessing, threshold).Find(&logs).Error
	return logs, err
}

func (r *bulkUploadLogRepository) FindByJobID(jobID string) (*models.BulkUploadLog, error) {
	var log models.BulkUploadLog
	err := r.db.Where("job_id = ?", jobID).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

