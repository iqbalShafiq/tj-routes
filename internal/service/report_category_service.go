package service

import (
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type ReportCategoryService interface {
	ListCategories() ([]models.ReportCategory, error)
	GetCategoryByName(name string) (*models.ReportCategory, error)
	CreateCategory(category *models.ReportCategory) error
}

type reportCategoryService struct {
	categoryRepo repository.ReportCategoryRepository
}

func NewReportCategoryService(categoryRepo repository.ReportCategoryRepository) ReportCategoryService {
	return &reportCategoryService{
		categoryRepo: categoryRepo,
	}
}

func (s *reportCategoryService) ListCategories() ([]models.ReportCategory, error) {
	return s.categoryRepo.List()
}

func (s *reportCategoryService) GetCategoryByName(name string) (*models.ReportCategory, error) {
	return s.categoryRepo.FindByName(name)
}

func (s *reportCategoryService) CreateCategory(category *models.ReportCategory) error {
	return s.categoryRepo.Create(category)
}

