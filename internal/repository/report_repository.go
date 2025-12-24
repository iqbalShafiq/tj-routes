package repository

import (
	"time"
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type ReportRepository interface {
	Create(report *models.Report) error
	FindByID(id uint) (*models.Report, error)
	Update(report *models.Report) error
	Delete(id uint) error
	List(offset, limit int, filters map[string]interface{}) ([]models.Report, int64, error)
	GetFeed(offset, limit int, filters map[string]interface{}, sort string) ([]models.Report, int64, error)
	GetTrending(offset, limit int, timeWindow string) ([]models.Report, int64, error)
	GetStories(userID *uint, limit int) ([]models.Report, error)
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

	// Sort parameter for enhanced sorting
	sort := "recent" // default
	if sortParam, ok := filters["sort"].(string); ok && sortParam != "" {
		sort = sortParam
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply ordering based on sort parameter
	switch sort {
	case "popular":
		query = query.Order("(upvotes - downvotes) DESC, created_at DESC")
	case "comments":
		query = query.Order("comment_count DESC, created_at DESC")
	case "trending":
		// For trending, we need to calculate a score
		// This is a simplified version - full trending logic is in GetTrending
		query = query.Order("((upvotes - downvotes) * 2 + (comment_count * 0.5)) DESC, created_at DESC")
	default: // recent
		query = query.Order("created_at DESC")
	}

	err := query.Preload("User").
		Preload("RelatedRoute").
		Preload("RelatedStop").
		Offset(offset).
		Limit(limit).
		Find(&reports).Error
	return reports, total, err
}

func (r *reportRepository) GetFeed(offset, limit int, filters map[string]interface{}, sort string) ([]models.Report, int64, error) {
	var reports []models.Report
	var total int64

	query := r.db.Model(&models.Report{})

	// Filter by followed users if provided
	if followedUserIDs, ok := filters["followed_user_ids"].([]uint); ok && len(followedUserIDs) > 0 {
		query = query.Where("user_id IN ?", followedUserIDs)
	}

	// Filter by hashtag if provided
	if hashtagID, ok := filters["hashtag_id"].(uint); ok && hashtagID > 0 {
		query = query.Joins("INNER JOIN report_hashtags ON reports.id = report_hashtags.report_id").
			Where("report_hashtags.hashtag_id = ?", hashtagID)
	}

	// Status filter (default: show all except deleted)
	if status, ok := filters["status"].(models.ReportStatus); ok {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	switch sort {
	case "popular":
		query = query.Order("(upvotes - downvotes) DESC, created_at DESC")
	case "trending":
		// Simplified trending - full calculation in GetTrending
		query = query.Order("((upvotes - downvotes) * 2 + (comment_count * 0.5)) DESC, created_at DESC")
	default: // recent
		query = query.Order("created_at DESC")
	}

	err := query.Preload("User").
		Preload("RelatedRoute").
		Preload("RelatedStop").
		Preload("Hashtags.Hashtag").
		Offset(offset).
		Limit(limit).
		Find(&reports).Error
	return reports, total, err
}

func (r *reportRepository) GetTrending(offset, limit int, timeWindow string) ([]models.Report, int64, error) {
	var reports []models.Report
	var total int64
	var timeThreshold time.Time
	now := time.Now()

	// Calculate time threshold based on window
	switch timeWindow {
	case "1h":
		timeThreshold = now.Add(-1 * time.Hour)
	case "24h":
		timeThreshold = now.Add(-24 * time.Hour)
	case "7d":
		timeThreshold = now.Add(-7 * 24 * time.Hour)
	case "30d":
		timeThreshold = now.Add(-30 * 24 * time.Hour)
	default: // "all"
		timeThreshold = time.Time{} // No time limit
	}

	query := r.db.Model(&models.Report{})

	if !timeThreshold.IsZero() {
		query = query.Where("created_at >= ?", timeThreshold)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Calculate trending score with age penalty and recency boost
	// score = (upvotes - downvotes) * 2 + (comment_count * 0.5) - age_penalty + recency_boost
	// We'll use SQL expressions for this
	err := query.
		Select("*, "+
			"((upvotes - downvotes) * 2 + "+
			"(comment_count * 0.5) - "+
			"(EXTRACT(EPOCH FROM (NOW() - created_at)) / 3600) * 0.1 + "+
			"CASE "+
			"WHEN created_at > NOW() - INTERVAL '1 hour' THEN 10 "+
			"WHEN created_at > NOW() - INTERVAL '24 hours' THEN 5 "+
			"ELSE 0 "+
			"END as trending_score").
		Order("trending_score DESC, created_at DESC").
		Preload("User").
		Preload("RelatedRoute").
		Preload("RelatedStop").
		Preload("Hashtags.Hashtag").
		Offset(offset).
		Limit(limit).
		Find(&reports).Error

	return reports, total, err
}

func (r *reportRepository) GetStories(userID *uint, limit int) ([]models.Report, error) {
	var reports []models.Report

	// Stories are recent reports with photos (last 24 hours)
	query := r.db.Model(&models.Report{}).
		Where("created_at >= ? AND photo_urls IS NOT NULL AND photo_urls != '[]' AND photo_urls != 'null'", time.Now().Add(-24*time.Hour))

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	err := query.Order("created_at DESC").
		Preload("User").
		Limit(limit).
		Find(&reports).Error

	return reports, err
}
