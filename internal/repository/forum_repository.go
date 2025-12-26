package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type ForumRepository interface {
	Create(forum *models.Forum) error
	FindByID(id uint) (*models.Forum, error)
	FindByRouteID(routeID uint) (*models.Forum, error)
	FindOrCreateByRouteID(routeID uint) (*models.Forum, error)
	Update(forum *models.Forum) error
	Delete(id uint) error
	GetMemberCount(forumID uint) (int64, error)
	IsMember(forumID uint, userID uint) (bool, error)
}

type forumRepository struct {
	db *gorm.DB
}

func NewForumRepository(db *gorm.DB) ForumRepository {
	return &forumRepository{db: db}
}

func (r *forumRepository) Create(forum *models.Forum) error {
	return r.db.Create(forum).Error
}

func (r *forumRepository) FindByID(id uint) (*models.Forum, error) {
	var forum models.Forum
	err := r.db.Preload("Route").First(&forum, id).Error
	if err != nil {
		return nil, err
	}
	return &forum, nil
}

func (r *forumRepository) FindByRouteID(routeID uint) (*models.Forum, error) {
	var forum models.Forum
	err := r.db.Where("route_id = ?", routeID).Preload("Route").First(&forum).Error
	if err != nil {
		return nil, err
	}
	return &forum, nil
}

func (r *forumRepository) FindOrCreateByRouteID(routeID uint) (*models.Forum, error) {
	forum, err := r.FindByRouteID(routeID)
	if err == nil {
		return forum, nil
	}

	// Forum doesn't exist, create it
	newForum := &models.Forum{
		RouteID: routeID,
	}
	if err := r.Create(newForum); err != nil {
		return nil, err
	}

	return r.FindByID(newForum.ID)
}

func (r *forumRepository) Update(forum *models.Forum) error {
	return r.db.Save(forum).Error
}

func (r *forumRepository) Delete(id uint) error {
	return r.db.Delete(&models.Forum{}, id).Error
}

func (r *forumRepository) GetMemberCount(forumID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.ForumMember{}).Where("forum_id = ?", forumID).Count(&count).Error
	return count, err
}

func (r *forumRepository) IsMember(forumID uint, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.ForumMember{}).
		Where("forum_id = ? AND user_id = ?", forumID, userID).
		Count(&count).Error
	return count > 0, err
}

