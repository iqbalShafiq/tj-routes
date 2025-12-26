package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type ForumPostRepository interface {
	Create(post *models.ForumPost) error
	FindByID(id uint) (*models.ForumPost, error)
	Update(post *models.ForumPost) error
	Delete(id uint) error
	ListByForumID(forumID uint, offset, limit int, filters map[string]interface{}) ([]models.ForumPost, int64, error)
	CountByForumID(forumID uint) (int64, error)
	IncrementCommentCount(postID uint) error
	GetRecentByForumID(forumID uint, limit int) ([]models.ForumPost, error)
}

type forumPostRepository struct {
	db *gorm.DB
}

func NewForumPostRepository(db *gorm.DB) ForumPostRepository {
	return &forumPostRepository{db: db}
}

func (r *forumPostRepository) Create(post *models.ForumPost) error {
	return r.db.Create(post).Error
}

func (r *forumPostRepository) FindByID(id uint) (*models.ForumPost, error) {
	var post models.ForumPost
	err := r.db.Preload("User").
		Preload("Forum").
		Preload("LinkedReport").
		First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *forumPostRepository) Update(post *models.ForumPost) error {
	return r.db.Save(post).Error
}

func (r *forumPostRepository) Delete(id uint) error {
	return r.db.Delete(&models.ForumPost{}, id).Error
}

func (r *forumPostRepository) ListByForumID(forumID uint, offset, limit int, filters map[string]interface{}) ([]models.ForumPost, int64, error) {
	var posts []models.ForumPost
	var total int64

	query := r.db.Model(&models.ForumPost{}).Where("forum_id = ?", forumID)

	// Filter by post type
	if postType, ok := filters["post_type"].(models.PostType); ok {
		query = query.Where("post_type = ?", postType)
	}

	// Search
	if search, ok := filters["search"].(string); ok && search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("(title ILIKE ? OR content ILIKE ?)", searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Order: pinned first, then by created_at desc
	err := query.Preload("User").
		Preload("LinkedReport").
		Order("is_pinned DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&posts).Error

	return posts, total, err
}

func (r *forumPostRepository) CountByForumID(forumID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.ForumPost{}).Where("forum_id = ?", forumID).Count(&count).Error
	return count, err
}

func (r *forumPostRepository) IncrementCommentCount(postID uint) error {
	return r.db.Model(&models.ForumPost{}).
		Where("id = ?", postID).
		UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
}

func (r *forumPostRepository) GetRecentByForumID(forumID uint, limit int) ([]models.ForumPost, error) {
	var posts []models.ForumPost

	err := r.db.Where("forum_id = ? AND deleted_at IS NULL", forumID).
		Order("created_at DESC").
		Limit(limit).
		Find(&posts).Error

	return posts, err
}

