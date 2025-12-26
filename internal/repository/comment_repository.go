package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(comment *models.Comment) error
	FindByID(id uint) (*models.Comment, error)
	FindByReportID(reportID uint) ([]models.Comment, error)
	FindThreadedByReportID(reportID uint) ([]models.Comment, error)
	FindByForumPostID(forumPostID uint) ([]models.Comment, error)
	FindThreadedByForumPostID(forumPostID uint) ([]models.Comment, error)
	Update(comment *models.Comment) error
	Delete(id uint) error
	CountByReportID(reportID uint) (int64, error)
	CountByForumPostID(forumPostID uint) (int64, error)
	CountByUserID(userID uint) (int64, error)
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

func (r *commentRepository) FindByID(id uint) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.Preload("User").First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepository) FindByReportID(reportID uint) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.Where("report_id = ?", reportID).
		Where("parent_id IS NULL"). // Only top-level comments
		Preload("User").
		Preload("Replies.User").
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

func (r *commentRepository) FindThreadedByReportID(reportID uint) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.Where("report_id = ?", reportID).
		Where("parent_id IS NULL"). // Only top-level comments
		Preload("User").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User").Order("created_at ASC")
		}).
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

func (r *commentRepository) Update(comment *models.Comment) error {
	return r.db.Save(comment).Error
}

func (r *commentRepository) Delete(id uint) error {
	return r.db.Delete(&models.Comment{}, id).Error
}

func (r *commentRepository) CountByReportID(reportID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Comment{}).Where("report_id = ?", reportID).Count(&count).Error
	return count, err
}

func (r *commentRepository) FindByForumPostID(forumPostID uint) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.Where("forum_post_id = ?", forumPostID).
		Where("parent_id IS NULL"). // Only top-level comments
		Preload("User").
		Preload("Replies.User").
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

func (r *commentRepository) FindThreadedByForumPostID(forumPostID uint) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.Where("forum_post_id = ?", forumPostID).
		Where("parent_id IS NULL"). // Only top-level comments
		Preload("User").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User").Order("created_at ASC")
		}).
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

func (r *commentRepository) CountByForumPostID(forumPostID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Comment{}).Where("forum_post_id = ?", forumPostID).Count(&count).Error
	return count, err
}

func (r *commentRepository) CountByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Comment{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

