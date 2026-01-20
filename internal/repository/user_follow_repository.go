package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type UserFollowRepository interface {
	Create(userFollow *models.UserFollow) error
	Delete(followerID, followingID uint) error
	IsFollowing(followerID, followingID uint) (bool, error)
	GetFollowers(userID uint, offset, limit int) ([]models.User, int64, error)
	GetFollowing(userID uint, offset, limit int) ([]models.User, int64, error)
	GetFollowCounts(userID uint) (followers int64, following int64, err error)
	GetFollowedUserIDs(followerID uint) ([]uint, error)
	GetFriends(userID uint, offset, limit int) ([]models.User, int64, error)
	IsMutualFollow(userA, userB uint) (bool, error)
}

type userFollowRepository struct {
	db *gorm.DB
}

func NewUserFollowRepository(db *gorm.DB) UserFollowRepository {
	return &userFollowRepository{db: db}
}

func (r *userFollowRepository) Create(userFollow *models.UserFollow) error {
	return r.db.Create(userFollow).Error
}

func (r *userFollowRepository) Delete(followerID, followingID uint) error {
	return r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(&models.UserFollow{}).Error
}

func (r *userFollowRepository) IsFollowing(followerID, followingID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.UserFollow{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count).Error
	return count > 0, err
}

func (r *userFollowRepository) GetFollowers(userID uint, offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// Count total followers
	err := r.db.Model(&models.UserFollow{}).
		Where("following_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get followers with pagination
	err = r.db.Model(&models.User{}).
		Joins("INNER JOIN user_follows ON users.id = user_follows.follower_id").
		Where("user_follows.following_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Find(&users).Error

	return users, total, err
}

func (r *userFollowRepository) GetFollowing(userID uint, offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// Count total following
	err := r.db.Model(&models.UserFollow{}).
		Where("follower_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get following with pagination
	err = r.db.Model(&models.User{}).
		Joins("INNER JOIN user_follows ON users.id = user_follows.following_id").
		Where("user_follows.follower_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Find(&users).Error

	return users, total, err
}

func (r *userFollowRepository) GetFollowCounts(userID uint) (followers int64, following int64, err error) {
	err = r.db.Model(&models.UserFollow{}).
		Where("following_id = ?", userID).
		Count(&followers).Error
	if err != nil {
		return 0, 0, err
	}

	err = r.db.Model(&models.UserFollow{}).
		Where("follower_id = ?", userID).
		Count(&following).Error
	if err != nil {
		return 0, 0, err
	}

	return followers, following, nil
}

func (r *userFollowRepository) GetFollowedUserIDs(followerID uint) ([]uint, error) {
	var userIDs []uint
	err := r.db.Model(&models.UserFollow{}).
		Where("follower_id = ?", followerID).
		Pluck("following_id", &userIDs).Error
	return userIDs, err
}

func (r *userFollowRepository) GetFriends(userID uint, offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// Count total friends (mutual follows)
	err := r.db.Model(&models.UserFollow{}).
		Joins("INNER JOIN user_follows uf2 ON user_follows.following_id = uf2.follower_id AND user_follows.follower_id = uf2.following_id").
		Where("user_follows.follower_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get friends with pagination
	err = r.db.Model(&models.User{}).
		Joins("INNER JOIN user_follows uf1 ON users.id = uf1.following_id").
		Joins("INNER JOIN user_follows uf2 ON uf1.following_id = uf2.follower_id AND uf1.follower_id = uf2.following_id").
		Where("uf1.follower_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Find(&users).Error

	return users, total, err
}

func (r *userFollowRepository) IsMutualFollow(userA, userB uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.UserFollow{}).
		Joins("INNER JOIN user_follows uf2 ON user_follows.following_id = uf2.follower_id AND user_follows.follower_id = uf2.following_id").
		Where("user_follows.follower_id = ? AND user_follows.following_id = ?", userA, userB).
		Count(&count).Error
	return count > 0, err
}

