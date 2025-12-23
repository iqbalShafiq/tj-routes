package service

import (
	"errors"
	"fmt"

	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type ReportService interface {
	CreateReport(report *models.Report) error
	GetReportByID(id uint) (*models.Report, error)
	UpdateReportStatus(id uint, status models.ReportStatus, adminNotes *string) error
	DeleteReport(id uint) error
	ListReports(offset, limit int, filters map[string]interface{}) ([]models.Report, int64, error)
}

type reportService struct {
	reportRepo        repository.ReportRepository
	routeRepo         repository.RouteRepository
	stopRepo          repository.StopRepository
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
	return s.reportRepo.Create(report)
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
