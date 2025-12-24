package service

import (
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type HashtagService interface {
	GetTrending(limit int) ([]models.Hashtag, error)
	GetReportsByHashtag(hashtagName string, offset, limit int) ([]models.Report, int64, error)
	SearchHashtags(query string, limit int) ([]models.Hashtag, error)
}

type hashtagService struct {
	hashtagRepo repository.HashtagRepository
}

func NewHashtagService(hashtagRepo repository.HashtagRepository) HashtagService {
	return &hashtagService{
		hashtagRepo: hashtagRepo,
	}
}

func (s *hashtagService) GetTrending(limit int) ([]models.Hashtag, error) {
	return s.hashtagRepo.GetTrending(limit)
}

func (s *hashtagService) GetReportsByHashtag(hashtagName string, offset, limit int) ([]models.Report, int64, error) {
	return s.hashtagRepo.GetReportsByHashtagName(hashtagName, offset, limit)
}

func (s *hashtagService) SearchHashtags(query string, limit int) ([]models.Hashtag, error) {
	return s.hashtagRepo.SearchHashtags(query, limit)
}

