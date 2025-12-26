package service

import (
	"encoding/json"
	"errors"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type ForumPostService interface {
	CreatePost(post *models.ForumPost, userID uint) error
	GetPostByID(postID uint) (*models.ForumPost, error)
	ListPosts(forumID uint, offset, limit int, filters map[string]interface{}) ([]models.ForumPost, int64, error)
	UpdatePost(postID uint, userID uint, title, content string, postType *models.PostType) error
	UpdatePostWithFiles(postID uint, userID uint, photoURLs, pdfURLs []string) error
	DeletePost(postID uint, userID uint, isAdmin bool) error
	PinPost(postID uint) error
	UnpinPost(postID uint) error
}

type forumPostService struct {
	forumPostRepo  repository.ForumPostRepository
	forumRepo      repository.ForumRepository
	forumMemberRepo repository.ForumMemberRepository
	reportRepo     repository.ReportRepository
}

func NewForumPostService(
	forumPostRepo repository.ForumPostRepository,
	forumRepo repository.ForumRepository,
	forumMemberRepo repository.ForumMemberRepository,
	reportRepo repository.ReportRepository,
) ForumPostService {
	return &forumPostService{
		forumPostRepo:  forumPostRepo,
		forumRepo:      forumRepo,
		forumMemberRepo: forumMemberRepo,
		reportRepo:     reportRepo,
	}
}

func (s *forumPostService) CreatePost(post *models.ForumPost, userID uint) error {
	// Verify forum exists
	_, err := s.forumRepo.FindByID(post.ForumID)
	if err != nil {
		return errors.New("forum not found")
	}

	// Check if user is a member (required to post)
	isMember, err := s.forumMemberRepo.IsMember(post.ForumID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("must be a forum member to create posts")
	}

	// Verify linked report exists if provided
	if post.LinkedReportID != nil {
		_, err := s.reportRepo.FindByID(*post.LinkedReportID)
		if err != nil {
			return errors.New("linked report not found")
		}
	}

	post.UserID = userID

	return s.forumPostRepo.Create(post)
}

func (s *forumPostService) GetPostByID(postID uint) (*models.ForumPost, error) {
	post, err := s.forumPostRepo.FindByID(postID)
	if err != nil {
		return nil, errors.New("forum post not found")
	}
	return post, nil
}

func (s *forumPostService) ListPosts(forumID uint, offset, limit int, filters map[string]interface{}) ([]models.ForumPost, int64, error) {
	// Verify forum exists
	_, err := s.forumRepo.FindByID(forumID)
	if err != nil {
		return nil, 0, errors.New("forum not found")
	}

	return s.forumPostRepo.ListByForumID(forumID, offset, limit, filters)
}

func (s *forumPostService) UpdatePost(postID uint, userID uint, title, content string, postType *models.PostType) error {
	post, err := s.forumPostRepo.FindByID(postID)
	if err != nil {
		return errors.New("forum post not found")
	}

	// Verify ownership
	if post.UserID != userID {
		return errors.New("unauthorized: you can only edit your own posts")
	}

	post.Title = title
	post.Content = content
	if postType != nil {
		post.PostType = *postType
	}

	return s.forumPostRepo.Update(post)
}

func (s *forumPostService) UpdatePostWithFiles(postID uint, userID uint, photoURLs, pdfURLs []string) error {
	post, err := s.forumPostRepo.FindByID(postID)
	if err != nil {
		return errors.New("forum post not found")
	}

	// Verify ownership
	if post.UserID != userID {
		return errors.New("unauthorized: you can only edit your own posts")
	}

	// Update file URLs
	if len(photoURLs) > 0 {
		photoData, _ := json.Marshal(photoURLs)
		photoStr := string(photoData)
		post.PhotoURLs = &photoStr
	}
	if len(pdfURLs) > 0 {
		pdfData, _ := json.Marshal(pdfURLs)
		pdfStr := string(pdfData)
		post.PDFURLs = &pdfStr
	}

	return s.forumPostRepo.Update(post)
}

func (s *forumPostService) DeletePost(postID uint, userID uint, isAdmin bool) error {
	post, err := s.forumPostRepo.FindByID(postID)
	if err != nil {
		return errors.New("forum post not found")
	}

	// Verify ownership or admin
	if post.UserID != userID && !isAdmin {
		return errors.New("unauthorized: you can only delete your own posts")
	}

	return s.forumPostRepo.Delete(postID)
}

func (s *forumPostService) PinPost(postID uint) error {
	post, err := s.forumPostRepo.FindByID(postID)
	if err != nil {
		return errors.New("forum post not found")
	}

	post.IsPinned = true
	return s.forumPostRepo.Update(post)
}

func (s *forumPostService) UnpinPost(postID uint) error {
	post, err := s.forumPostRepo.FindByID(postID)
	if err != nil {
		return errors.New("forum post not found")
	}

	post.IsPinned = false
	return s.forumPostRepo.Update(post)
}

// Helper function to serialize photo/PDF URLs to JSON
func serializeURLs(urls []string) *string {
	if len(urls) == 0 {
		return nil
	}
	data, _ := json.Marshal(urls)
	result := string(data)
	return &result
}

// Helper function to deserialize photo/PDF URLs from JSON
func deserializeURLs(data *string) []string {
	if data == nil || *data == "" {
		return []string{}
	}
	var urls []string
	json.Unmarshal([]byte(*data), &urls)
	return urls
}

