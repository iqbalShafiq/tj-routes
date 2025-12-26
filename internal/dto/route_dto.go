package dto

import (
	"time"

	"tj-routes/internal/models"
)

type RouteDetailResponse struct {
	Route         *models.Route      `json:"route"`
	Statistics    RouteStatistics    `json:"statistics"`
	RecentReports []ReportSummary    `json:"recent_reports,omitempty"`
	RecentPosts   []ForumPostSummary `json:"recent_posts,omitempty"`
}

type RouteStatistics struct {
	// Forum reference
	ForumID *uint `json:"forum_id,omitempty"`

	// Report stats
	TotalReports        int            `json:"total_reports"`
	ReportsByStatus     map[string]int `json:"reports_by_status"`     // pending, reviewed, resolved
	ReportsByType       map[string]int `json:"reports_by_type"`       // route_issue, stop_issue, etc.
	TotalReportUpvotes  int            `json:"total_report_upvotes"`
	TotalReportDownvotes int           `json:"total_report_downvotes"`

	// Forum stats
	ForumMemberCount   int `json:"forum_member_count"`
	ForumPostCount     int `json:"forum_post_count"`
	TotalPostUpvotes   int `json:"total_post_upvotes"`
	TotalPostDownvotes int `json:"total_post_downvotes"`

	// Computed metrics
	ReportResolutionRate   float64 `json:"report_resolution_rate"`   // resolved / total
	CommunityActivityScore float64 `json:"community_activity_score"` // weighted score
}

type ReportSummary struct {
	ID        uint      `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Upvotes   int       `json:"upvotes"`
	CreatedAt time.Time `json:"created_at"`
}

type ForumPostSummary struct {
	ID        uint      `json:"id"`
	PostType  string    `json:"post_type"`
	Title     string    `json:"title"`
	Upvotes   int       `json:"upvotes"`
	CreatedAt time.Time `json:"created_at"`
}
