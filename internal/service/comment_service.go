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
	GetCommentsByForumPostID(forumPostID uint) ([]models.Comment, error)
	UpdateComment(id uint, userID uint, content string) error
	DeleteComment(id uint, userID uint) error
}

type commentService struct {
	commentRepo    repository.CommentRepository
	reportRepo     repository.ReportRepository
	forumPostRepo  repository.ForumPostRepository
}

func NewCommentService(commentRepo repository.CommentRepository, reportRepo repository.ReportRepository) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		reportRepo:  reportRepo,
	}
}

func NewCommentServiceWithForumPost(commentRepo repository.CommentRepository, reportRepo repository.ReportRepository, forumPostRepo repository.ForumPostRepository) CommentService {
	return &commentService{
		commentRepo:   commentRepo,
		reportRepo:    reportRepo,
		forumPostRepo: forumPostRepo,
	}
}

func (s *commentService) CreateComment(comment *models.Comment) error {
	// Verify either report or forum post exists
	if comment.ReportID != nil {
		_, err := s.reportRepo.FindByID(*comment.ReportID)
		if err != nil {
			return errors.New("report not found")
		}
	} else if comment.ForumPostID != nil {
		if s.forumPostRepo == nil {
			return errors.New("forum post repository not available")
		}
		_, err := s.forumPostRepo.FindByID(*comment.ForumPostID)
		if err != nil {
			return errors.New("forum post not found")
		}
	} else {
		return errors.New("comment must be associated with either a report or forum post")
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

	// Update comment count
	if comment.ReportID != nil {
		report, err := s.reportRepo.FindByID(*comment.ReportID)
		if err == nil {
			count, _ := s.commentRepo.CountByReportID(*comment.ReportID)
			report.CommentCount = int(count)
			s.reportRepo.Update(report)
		}
	} else if comment.ForumPostID != nil && s.forumPostRepo != nil {
		forumPost, err := s.forumPostRepo.FindByID(*comment.ForumPostID)
		if err == nil {
			return nil // Continue even if update fails
		}
		count, _ := s.commentRepo.CountByForumPostID(*comment.ForumPostID)
		forumPost.CommentCount = int(count)
		s.forumPostRepo.Update(forumPost)
	}

	return nil
}

func (s *commentService) GetCommentByID(id uint) (*models.Comment, error) {
	return s.commentRepo.FindByID(id)
}

func (s *commentService) GetCommentsByReportID(reportID uint) ([]models.Comment, error) {
	return s.commentRepo.FindThreadedByReportID(reportID)
}

func (s *commentService) GetCommentsByForumPostID(forumPostID uint) ([]models.Comment, error) {
	if s.forumPostRepo == nil {
		return nil, errors.New("forum post repository not available")
	}
	return s.commentRepo.FindThreadedByForumPostID(forumPostID)
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

