package repository

import (
	"database/sql"

	"tj-routes/internal/dto"
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type RouteRepository interface {
	Create(route *models.Route) error
	FindByID(id uint) (*models.Route, error)
	FindByRouteNumber(routeNumber string) (*models.Route, error)
	Update(route *models.Route) error
	Delete(id uint) error
	List(offset, limit int, filters map[string]interface{}) ([]models.Route, int64, error)
	GetRouteWithStats(id uint) (*models.Route, *dto.RouteStatistics, error)
}

type routeRepository struct {
	db *gorm.DB
}

func NewRouteRepository(db *gorm.DB) RouteRepository {
	return &routeRepository{db: db}
}

func (r *routeRepository) Create(route *models.Route) error {
	return r.db.Create(route).Error
}

func (r *routeRepository) FindByID(id uint) (*models.Route, error) {
	var route models.Route
	err := r.db.Preload("RouteStops.Stop").First(&route, id).Error
	if err != nil {
		return nil, err
	}
	return &route, nil
}

func (r *routeRepository) FindByRouteNumber(routeNumber string) (*models.Route, error) {
	var route models.Route
	err := r.db.Where("LOWER(route_number) = LOWER(?)", routeNumber).First(&route).Error
	if err != nil {
		return nil, err
	}
	return &route, nil
}

func (r *routeRepository) Update(route *models.Route) error {
	return r.db.Save(route).Error
}

func (r *routeRepository) Delete(id uint) error {
	return r.db.Delete(&models.Route{}, id).Error
}

func (r *routeRepository) List(offset, limit int, filters map[string]interface{}) ([]models.Route, int64, error) {
	var routes []models.Route
	var total int64

	query := r.db.Model(&models.Route{})

	if status, ok := filters["status"].(models.Status); ok {
		query = query.Where("status = ?", status)
	}
	if routeNumber, ok := filters["route_number"].(string); ok && routeNumber != "" {
		query = query.Where("route_number = ?", routeNumber)
	}

	// Fuzzy search using pg_trgm with similarity threshold and ILIKE for partial matching
	if search, ok := filters["search"].(string); ok && search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"(similarity(name, ?) > 0.2 OR similarity(route_number, ?) > 0.2 OR similarity(description, ?) > 0.2) OR (name ILIKE ? OR route_number ILIKE ? OR description ILIKE ?)",
			search, search, search, searchPattern, searchPattern, searchPattern,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Note: We skip relevance-based ordering for now due to GORM limitations with parameterized ORDER BY
	// The search filtering still works correctly via the WHERE clause
	// Results will be ordered by created_at (default) or can be ordered by other fields if needed

	err := query.Preload("RouteStops.Stop").Offset(offset).Limit(limit).Find(&routes).Error
	return routes, total, err
}

func (r *routeRepository) GetRouteWithStats(id uint) (*models.Route, *dto.RouteStatistics, error) {
	var route models.Route
	err := r.db.Preload("RouteStops.Stop").First(&route, id).Error
	if err != nil {
		return nil, nil, err
	}

	stats := &dto.RouteStatistics{
		ReportsByStatus: make(map[string]int),
		ReportsByType:   make(map[string]int),
	}

	// Get forum_id first
	var forumID uint
	err = r.db.Model(&models.Forum{}).Select("id").Where("route_id = ?", id).Scan(&forumID).Error
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, err
	}
	if forumID > 0 {
		stats.ForumID = &forumID
	}

	// Get report statistics using raw SQL with subqueries
	reportQuery := `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE status = 'reviewed') as reviewed,
			COUNT(*) FILTER (WHERE status = 'resolved') as resolved,
			COUNT(*) FILTER (WHERE type = 'route_issue') as route_issue_count,
			COUNT(*) FILTER (WHERE type = 'stop_issue') as stop_issue_count,
			COUNT(*) FILTER (WHERE type = 'temporary_event') as temporary_event_count,
			COUNT(*) FILTER (WHERE type = 'policy_change') as policy_change_count,
			COALESCE(SUM(upvotes), 0) as total_upvotes,
			COALESCE(SUM(downvotes), 0) as total_downvotes
		FROM reports
		WHERE related_route_id = ? AND deleted_at IS NULL
	`
	var total, pending, reviewed, resolved int64
	var routeIssueCount, stopIssueCount, temporaryEventCount, policyChangeCount int64
	var totalUpvotes, totalDownvotes int

	err = r.db.Raw(reportQuery, id).Row().Scan(
		&total, &pending, &reviewed, &resolved,
		&routeIssueCount, &stopIssueCount, &temporaryEventCount, &policyChangeCount,
		&totalUpvotes, &totalDownvotes,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, err
	}

	stats.TotalReports = int(total)
	stats.ReportsByStatus["pending"] = int(pending)
	stats.ReportsByStatus["reviewed"] = int(reviewed)
	stats.ReportsByStatus["resolved"] = int(resolved)
	stats.ReportsByType["route_issue"] = int(routeIssueCount)
	stats.ReportsByType["stop_issue"] = int(stopIssueCount)
	stats.ReportsByType["temporary_event"] = int(temporaryEventCount)
	stats.ReportsByType["policy_change"] = int(policyChangeCount)
	stats.TotalReportUpvotes = totalUpvotes
	stats.TotalReportDownvotes = totalDownvotes

	// Get forum statistics
	if forumID > 0 {
		forumQuery := `
			SELECT
				COALESCE((SELECT COUNT(*) FROM forum_members WHERE forum_id = ?), 0) as member_count,
				COALESCE((SELECT COUNT(*) FROM forum_posts WHERE forum_id = ? AND deleted_at IS NULL), 0) as post_count,
				COALESCE((SELECT SUM(upvotes) FROM forum_posts WHERE forum_id = ? AND deleted_at IS NULL), 0) as post_upvotes,
				COALESCE((SELECT SUM(downvotes) FROM forum_posts WHERE forum_id = ? AND deleted_at IS NULL), 0) as post_downvotes
		`
		var memberCount, postCount, postUpvotes, postDownvotes int64
		err = r.db.Raw(forumQuery, forumID, forumID, forumID, forumID).Row().Scan(
			&memberCount, &postCount, &postUpvotes, &postDownvotes,
		)
		if err != nil && err != sql.ErrNoRows {
			return nil, nil, err
		}

		stats.ForumMemberCount = int(memberCount)
		stats.ForumPostCount = int(postCount)
		stats.TotalPostUpvotes = int(postUpvotes)
		stats.TotalPostDownvotes = int(postDownvotes)
	}

	// Calculate derived metrics
	if stats.TotalReports > 0 {
		stats.ReportResolutionRate = float64(stats.ReportsByStatus["resolved"]) / float64(stats.TotalReports)
	}

	// Community activity score: weighted formula
	// (posts * 1) + (members * 0.5) + (upvotes * 0.1) - (downvotes * 0.1)
	stats.CommunityActivityScore = float64(stats.ForumPostCount)*1.0 +
		float64(stats.ForumMemberCount)*0.5 +
		float64(stats.TotalPostUpvotes)*0.1 -
		float64(stats.TotalPostDownvotes)*0.1

	return &route, stats, nil
}

