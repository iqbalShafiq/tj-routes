package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type ReportCategoryRepository interface {
	List() ([]models.ReportCategory, error)
	FindByName(name string) (*models.ReportCategory, error)
	Create(category *models.ReportCategory) error
}

type reportCategoryRepository struct {
	db *gorm.DB
}

func NewReportCategoryRepository(db *gorm.DB) ReportCategoryRepository {
	return &reportCategoryRepository{db: db}
}

func (r *reportCategoryRepository) List() ([]models.ReportCategory, error) {
	var categories []models.ReportCategory
	err := r.db.Order("name ASC").Find(&categories).Error
	return categories, err
}

func (r *reportCategoryRepository) FindByName(name string) (*models.ReportCategory, error) {
	var category models.ReportCategory
	err := r.db.Where("name = ?", name).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *reportCategoryRepository) Create(category *models.ReportCategory) error {
	return r.db.Create(category).Error
}

