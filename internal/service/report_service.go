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
	reportRepo repository.ReportRepository
	routeRepo  repository.RouteRepository
	stopRepo   repository.StopRepository
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

	report.Status = status
	if adminNotes != nil {
		report.AdminNotes = adminNotes
	}

	return s.reportRepo.Update(report)
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
