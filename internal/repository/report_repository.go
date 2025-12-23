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

	// Fuzzy search using pg_trgm with similarity threshold and ILIKE for partial matching
	if search, ok := filters["search"].(string); ok && search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"(similarity(title, ?) > 0.2 OR similarity(description, ?) > 0.2) OR (title ILIKE ? OR description ILIKE ?)",
			search, search, searchPattern, searchPattern,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply ordering - created_at DESC for all (relevance ordering skipped due to GORM limitations)
	query = query.Order("created_at DESC")

	err := query.Preload("User").
		Preload("RelatedRoute").
		Preload("RelatedStop").
		Offset(offset).
		Limit(limit).
		Find(&reports).Error
	return reports, total, err
}
