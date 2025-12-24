package service

import (
	"errors"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type UserFollowService interface {
	Follow(followerID, followingID uint) error
	Unfollow(followerID, followingID uint) error
	GetFollowers(userID uint, offset, limit int) ([]models.User, int64, error)
	GetFollowing(userID uint, offset, limit int) ([]models.User, int64, error)
	GetFollowStats(userID uint) (followers int64, following int64, err error)
	IsFollowing(followerID, followingID uint) (bool, error)
	GetFollowedUserIDs(followerID uint) ([]uint, error)
}

type userFollowService struct {
	userFollowRepo repository.UserFollowRepository
	userRepo       repository.UserRepository
}

func NewUserFollowService(userFollowRepo repository.UserFollowRepository, userRepo repository.UserRepository) UserFollowService {
	return &userFollowService{
		userFollowRepo: userFollowRepo,
		userRepo:       userRepo,
	}
}

func (s *userFollowService) Follow(followerID, followingID uint) error {
	// Prevent self-follow
	if followerID == followingID {
		return errors.New("cannot follow yourself")
	}

	// Check if user exists
	_, err := s.userRepo.FindByID(followingID)
	if err != nil {
		return errors.New("user not found")
	}

	// Check if already following
	isFollowing, err := s.userFollowRepo.IsFollowing(followerID, followingID)
	if err != nil {
		return err
	}
	if isFollowing {
		return errors.New("already following this user")
	}

	userFollow := &models.UserFollow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	return s.userFollowRepo.Create(userFollow)
}

func (s *userFollowService) Unfollow(followerID, followingID uint) error {
	// Check if following
	isFollowing, err := s.userFollowRepo.IsFollowing(followerID, followingID)
	if err != nil {
		return err
	}
	if !isFollowing {
		return errors.New("not following this user")
	}

	return s.userFollowRepo.Delete(followerID, followingID)
}

func (s *userFollowService) GetFollowers(userID uint, offset, limit int) ([]models.User, int64, error) {
	return s.userFollowRepo.GetFollowers(userID, offset, limit)
}

func (s *userFollowService) GetFollowing(userID uint, offset, limit int) ([]models.User, int64, error) {
	return s.userFollowRepo.GetFollowing(userID, offset, limit)
}

func (s *userFollowService) GetFollowStats(userID uint) (followers int64, following int64, err error) {
	return s.userFollowRepo.GetFollowCounts(userID)
}

func (s *userFollowService) IsFollowing(followerID, followingID uint) (bool, error) {
	return s.userFollowRepo.IsFollowing(followerID, followingID)
}

func (s *userFollowService) GetFollowedUserIDs(followerID uint) ([]uint, error) {
	return s.userFollowRepo.GetFollowedUserIDs(followerID)
}

