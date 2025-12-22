package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type ReportRepository interface {
	Create(report *models.Report) error
	FindByID(id uint) (*models.Report, error)
	Update(report *models.Report) error
	Delete(id uint) error
	List(offset, limit int, filters map[string]interface{}) ([]models.Report, int64, error)
}

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) Create(report *models.Report) error {
	return r.db.Create(report).Error
}

func (r *reportRepository) FindByID(id uint) (*models.Report, error) {
	var report models.Report
	err := r.db.Preload("User").
		Preload("RelatedRoute").
		Preload("RelatedStop").
		First(&report, id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *reportRepository) Update(report *models.Report) error {
	return r.db.Save(report).Error
}

func (r *reportRepository) Delete(id uint) error {
	return r.db.Delete(&models.Report{}, id).Error
}

func (r *reportRepository) List(offset, limit int, filters map[string]interface{}) ([]models.Report, int64, error) {
	var reports []models.Report
	var total int64

	query := r.db.Model(&models.Report{})

	if status, ok := filters["status"].(models.ReportStatus); ok {
		query = query.Where("status = ?", status)
	}
	if userID, ok := filters["user_id"].(uint); ok && userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if reportType, ok := filters["type"].(models.ReportType); ok {
		query = query.Where("type = ?", reportType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("User").
		Preload("RelatedRoute").
		Preload("RelatedStop").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&reports).Error
	return reports, total, err
}
