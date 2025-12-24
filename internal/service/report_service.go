package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type ReportService interface {
	CreateReport(report *models.Report) error
	GetReportByID(id uint) (*models.Report, error)
	UpdateReportStatus(id uint, status models.ReportStatus, adminNotes *string) error
	DeleteReport(id uint) error
	ListReports(offset, limit int, filters map[string]interface{}) ([]models.Report, int64, error)
	GetFeed(offset, limit int, filters map[string]interface{}, sort string, userID *uint) ([]models.Report, int64, error)
	GetTrending(offset, limit int, timeWindow string) ([]models.Report, int64, error)
	GetStories(userID *uint, limit int) ([]models.Report, error)
	ExtractHashtags(text string) []string
	UpdateReportHashtags(reportID uint, hashtags []string) error
}

type reportService struct {
	reportRepo        repository.ReportRepository
	routeRepo         repository.RouteRepository
	stopRepo          repository.StopRepository
	hashtagRepo       repository.HashtagRepository
	userFollowService UserFollowService
	reputationService ReputationService
	badgeService      BadgeService
}

func NewReportService(
	reportRepo repository.ReportRepository,
	routeRepo repository.RouteRepository,
	stopRepo repository.StopRepository,
) ReportService {
	return &reportService{
		reportRepo: reportRepo,
		routeRepo:  routeRepo,
		stopRepo:   stopRepo,
	}
}

func NewReportServiceWithReputation(
	reportRepo repository.ReportRepository,
	routeRepo repository.RouteRepository,
	stopRepo repository.StopRepository,
	reputationService ReputationService,
	badgeService BadgeService,
) ReportService {
	return &reportService{
		reportRepo:        reportRepo,
		routeRepo:         routeRepo,
		stopRepo:          stopRepo,
		reputationService: reputationService,
		badgeService:      badgeService,
	}
}

func NewReportServiceWithSocial(
	reportRepo repository.ReportRepository,
	routeRepo repository.RouteRepository,
	stopRepo repository.StopRepository,
	hashtagRepo repository.HashtagRepository,
	userFollowService UserFollowService,
	reputationService ReputationService,
	badgeService BadgeService,
) ReportService {
	return &reportService{
		reportRepo:        reportRepo,
		routeRepo:         routeRepo,
		stopRepo:          stopRepo,
		hashtagRepo:       hashtagRepo,
		userFollowService: userFollowService,
		reputationService: reputationService,
		badgeService:      badgeService,
	}
}

func (s *reportService) CreateReport(report *models.Report) error {
	// Validate related route exists if provided
	if report.RelatedRouteID != nil {
		_, err := s.routeRepo.FindByID(*report.RelatedRouteID)
		if err != nil {
			return fmt.Errorf("related route not found: %w", err)
		}
	}

	// Validate related stop exists if provided
	if report.RelatedStopID != nil {
		_, err := s.stopRepo.FindByID(*report.RelatedStopID)
		if err != nil {
			return fmt.Errorf("related stop not found: %w", err)
		}
	}

	report.Status = models.ReportStatusPending
	
	// Create report first to get ID
	if err := s.reportRepo.Create(report); err != nil {
		return err
	}

	// Extract and save hashtags if hashtag repo is available
	if s.hashtagRepo != nil {
		text := report.Title + " " + report.Description
		hashtags := s.ExtractHashtags(text)
		if len(hashtags) > 0 {
			if err := s.UpdateReportHashtags(report.ID, hashtags); err != nil {
				// Log error but don't fail report creation
				// In production, you might want to log this
			}
		}
	}

	return nil
}

func (s *reportService) GetReportByID(id uint) (*models.Report, error) {
	return s.reportRepo.FindByID(id)
}

func (s *reportService) UpdateReportStatus(id uint, status models.ReportStatus, adminNotes *string) error {
	report, err := s.reportRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Validate status transition
	if !isValidStatusTransition(report.Status, status) {
		return errors.New("invalid status transition")
	}

	oldStatus := report.Status
	report.Status = status
	if adminNotes != nil {
		report.AdminNotes = adminNotes
	}

	if err := s.reportRepo.Update(report); err != nil {
		return err
	}

	// Award points based on status change
	if s.reputationService != nil {
		var pointsToAward int
		switch status {
		case models.ReportStatusReviewed:
			// Only award if transitioning from pending
			if oldStatus == models.ReportStatusPending {
				pointsToAward = 10
			}
		case models.ReportStatusResolved:
			// Award 25 points if transitioning to resolved
			// If already reviewed, award the difference (15 points)
			if oldStatus == models.ReportStatusPending {
				pointsToAward = 25
			} else if oldStatus == models.ReportStatusReviewed {
				pointsToAward = 15 // Total 25, already got 10
			}
		}

		if pointsToAward > 0 {
			if err := s.reputationService.AddPoints(report.UserID, pointsToAward); err == nil {
				// Check for badges after awarding points
				if s.badgeService != nil {
					s.badgeService.CheckAndAwardBadges(report.UserID)
				}
			}
		}
	}

	return nil
}

func (s *reportService) DeleteReport(id uint) error {
	return s.reportRepo.Delete(id)
}

func (s *reportService) ListReports(offset, limit int, filters map[string]interface{}) ([]models.Report, int64, error) {
	return s.reportRepo.List(offset, limit, filters)
}

func (s *reportService) GetFeed(offset, limit int, filters map[string]interface{}, sort string, userID *uint) ([]models.Report, int64, error) {
	// If filtering by followed users and user is authenticated
	if followed, ok := filters["followed"].(bool); ok && followed && userID != nil && s.userFollowService != nil {
		followedUserIDs, err := s.userFollowService.GetFollowedUserIDs(*userID)
		if err != nil {
			return nil, 0, err
		}
		if len(followedUserIDs) == 0 {
			// User follows no one, return empty result
			return []models.Report{}, 0, nil
		}
		filters["followed_user_ids"] = followedUserIDs
	}

	// Handle hashtag filter
	if hashtagName, ok := filters["hashtag"].(string); ok && hashtagName != "" && s.hashtagRepo != nil {
		hashtag, err := s.hashtagRepo.FindOrCreate(hashtagName)
		if err != nil {
			return nil, 0, err
		}
		filters["hashtag_id"] = hashtag.ID
	}

	return s.reportRepo.GetFeed(offset, limit, filters, sort)
}

func (s *reportService) GetTrending(offset, limit int, timeWindow string) ([]models.Report, int64, error) {
	return s.reportRepo.GetTrending(offset, limit, timeWindow)
}

func (s *reportService) GetStories(userID *uint, limit int) ([]models.Report, error) {
	return s.reportRepo.GetStories(userID, limit)
}

func (s *reportService) ExtractHashtags(text string) []string {
	// Regex to find hashtags: # followed by word characters
	hashtagRegex := regexp.MustCompile(`#[\w]+`)
	matches := hashtagRegex.FindAllString(text, -1)
	
	// Normalize hashtags (lowercase, remove duplicates)
	hashtagMap := make(map[string]bool)
	var hashtags []string
	
	for _, match := range matches {
		// Remove # and convert to lowercase
		name := strings.ToLower(strings.TrimPrefix(match, "#"))
		if name != "" && !hashtagMap[name] {
			hashtagMap[name] = true
			hashtags = append(hashtags, name)
		}
	}
	
	return hashtags
}

func (s *reportService) UpdateReportHashtags(reportID uint, hashtags []string) error {
	if s.hashtagRepo == nil {
		return errors.New("hashtag repository not available")
	}

	// Delete existing hashtags for this report
	if err := s.hashtagRepo.DeleteReportHashtags(reportID); err != nil {
		return err
	}

	// Create new hashtag associations
	for _, hashtagName := range hashtags {
		hashtag, err := s.hashtagRepo.FindOrCreate(hashtagName)
		if err != nil {
			continue // Skip on error, continue with others
		}

		// Create report-hashtag association
		if err := s.hashtagRepo.CreateReportHashtag(reportID, hashtag.ID); err != nil {
			continue
		}

		// Increment usage count
		s.hashtagRepo.IncrementUsageCount(hashtag.ID)
	}

	return nil
}

func isValidStatusTransition(current, new models.ReportStatus) bool {
	switch current {
	case models.ReportStatusPending:
		return new == models.ReportStatusReviewed || new == models.ReportStatusResolved
	case models.ReportStatusReviewed:
		return new == models.ReportStatusResolved
	case models.ReportStatusResolved:
		return false // Cannot change from resolved
	default:
		return false
	}
}
