package repository

import (
	"strings"
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type HashtagRepository interface {
	FindOrCreate(name string) (*models.Hashtag, error)
	GetTrending(limit int) ([]models.Hashtag, error)
	GetByReport(reportID uint) ([]models.Hashtag, error)
	GetReportsByHashtag(hashtagID uint, offset, limit int) ([]models.Report, int64, error)
	GetReportsByHashtagName(hashtagName string, offset, limit int) ([]models.Report, int64, error)
	SearchHashtags(query string, limit int) ([]models.Hashtag, error)
	CreateReportHashtag(reportID, hashtagID uint) error
	DeleteReportHashtags(reportID uint) error
	IncrementUsageCount(hashtagID uint) error
	DecrementUsageCount(hashtagID uint) error
}

type hashtagRepository struct {
	db *gorm.DB
}

func NewHashtagRepository(db *gorm.DB) HashtagRepository {
	return &hashtagRepository{db: db}
}

func (r *hashtagRepository) FindOrCreate(name string) (*models.Hashtag, error) {
	// Normalize hashtag name (lowercase, remove # if present)
	name = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(name), "#"))
	
	var hashtag models.Hashtag
	err := r.db.Where("name = ?", name).First(&hashtag).Error
	
	if err == gorm.ErrRecordNotFound {
		// Create new hashtag
		hashtag = models.Hashtag{
			Name:       name,
			UsageCount: 0,
		}
		if err := r.db.Create(&hashtag).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	
	return &hashtag, nil
}

func (r *hashtagRepository) GetTrending(limit int) ([]models.Hashtag, error) {
	var hashtags []models.Hashtag
	err := r.db.Order("usage_count DESC, updated_at DESC").
		Limit(limit).
		Find(&hashtags).Error
	return hashtags, err
}

func (r *hashtagRepository) GetByReport(reportID uint) ([]models.Hashtag, error) {
	var hashtags []models.Hashtag
	err := r.db.Model(&models.Hashtag{}).
		Joins("INNER JOIN report_hashtags ON hashtags.id = report_hashtags.hashtag_id").
		Where("report_hashtags.report_id = ?", reportID).
		Find(&hashtags).Error
	return hashtags, err
}

func (r *hashtagRepository) GetReportsByHashtag(hashtagID uint, offset, limit int) ([]models.Report, int64, error) {
	var reports []models.Report
	var total int64

	// Count total
	err := r.db.Model(&models.ReportHashtag{}).
		Where("hashtag_id = ?", hashtagID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get reports
	err = r.db.Model(&models.Report{}).
		Joins("INNER JOIN report_hashtags ON reports.id = report_hashtags.report_id").
		Where("report_hashtags.hashtag_id = ?", hashtagID).
		Preload("User").
		Preload("RelatedRoute").
		Preload("RelatedStop").
		Order("reports.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&reports).Error

	return reports, total, err
}

func (r *hashtagRepository) GetReportsByHashtagName(hashtagName string, offset, limit int) ([]models.Report, int64, error) {
	hashtagName = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(hashtagName), "#"))
	
	var hashtag models.Hashtag
	err := r.db.Where("name = ?", hashtagName).First(&hashtag).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return []models.Report{}, 0, nil
		}
		return nil, 0, err
	}

	return r.GetReportsByHashtag(hashtag.ID, offset, limit)
}

func (r *hashtagRepository) SearchHashtags(query string, limit int) ([]models.Hashtag, error) {
	var hashtags []models.Hashtag
	searchPattern := "%" + strings.ToLower(query) + "%"
	err := r.db.Where("name ILIKE ?", searchPattern).
		Order("usage_count DESC").
		Limit(limit).
		Find(&hashtags).Error
	return hashtags, err
}

func (r *hashtagRepository) CreateReportHashtag(reportID, hashtagID uint) error {
	reportHashtag := &models.ReportHashtag{
		ReportID:  reportID,
		HashtagID: hashtagID,
	}
	return r.db.Create(reportHashtag).Error
}

func (r *hashtagRepository) DeleteReportHashtags(reportID uint) error {
	return r.db.Where("report_id = ?", reportID).Delete(&models.ReportHashtag{}).Error
}

func (r *hashtagRepository) IncrementUsageCount(hashtagID uint) error {
	return r.db.Model(&models.Hashtag{}).
		Where("id = ?", hashtagID).
		Update("usage_count", gorm.Expr("usage_count + 1")).Error
}

func (r *hashtagRepository) DecrementUsageCount(hashtagID uint) error {
	return r.db.Model(&models.Hashtag{}).
		Where("id = ? AND usage_count > 0", hashtagID).
		Update("usage_count", gorm.Expr("usage_count - 1")).Error
}

