package service

import (
	"errors"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type CommentService interface {
	CreateComment(comment *models.Comment) error
	GetCommentByID(id uint) (*models.Comment, error)
	GetCommentsByReportID(reportID uint) ([]models.Comment, error)
	UpdateComment(id uint, userID uint, content string) error
	DeleteComment(id uint, userID uint) error
}

type commentService struct {
	commentRepo repository.CommentRepository
	reportRepo  repository.ReportRepository
}

func NewCommentService(commentRepo repository.CommentRepository, reportRepo repository.ReportRepository) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		reportRepo:  reportRepo,
	}
}

func (s *commentService) CreateComment(comment *models.Comment) error {
	// Verify report exists
	_, err := s.reportRepo.FindByID(comment.ReportID)
	if err != nil {
		return errors.New("report not found")
	}

	// If parent_id is set, verify parent comment exists
	if comment.ParentID != nil {
		_, err := s.commentRepo.FindByID(*comment.ParentID)
		if err != nil {
			return errors.New("parent comment not found")
		}
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return err
	}

	// Update report comment count
	report, err := s.reportRepo.FindByID(comment.ReportID)
	if err == nil {
		count, _ := s.commentRepo.CountByReportID(comment.ReportID)
		report.CommentCount = int(count)
		s.reportRepo.Update(report)
	}

	return nil
}

func (s *commentService) GetCommentByID(id uint) (*models.Comment, error) {
	return s.commentRepo.FindByID(id)
}

func (s *commentService) GetCommentsByReportID(reportID uint) ([]models.Comment, error) {
	return s.commentRepo.FindThreadedByReportID(reportID)
}

func (s *commentService) UpdateComment(id uint, userID uint, content string) error {
	comment, err := s.commentRepo.FindByID(id)
	if err != nil {
		return errors.New("comment not found")
	}

	// Verify ownership
	if comment.UserID != userID {
		return errors.New("unauthorized: you can only edit your own comments")
	}

	comment.Content = content
	return s.commentRepo.Update(comment)
}

func (s *commentService) DeleteComment(id uint, userID uint) error {
	comment, err := s.commentRepo.FindByID(id)
	if err != nil {
		return errors.New("comment not found")
	}

	// Verify ownership
	if comment.UserID != userID {
		return errors.New("unauthorized: you can only delete your own comments")
	}

	return s.commentRepo.Delete(id)
}

