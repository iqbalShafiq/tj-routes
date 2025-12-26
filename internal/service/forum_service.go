package service

import (
	"errors"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type ForumService interface {
	GetOrCreateForumByRouteID(routeID uint) (*models.Forum, error)
	GetForumByRouteID(routeID uint) (*models.Forum, error)
	GetForumByID(forumID uint) (*models.Forum, error)
	JoinForum(forumID uint, userID uint) error
	LeaveForum(forumID uint, userID uint) error
	IsMember(forumID uint, userID uint) (bool, error)
	GetForumMembers(forumID uint, offset, limit int) ([]models.ForumMember, int64, error)
	GetMemberCount(forumID uint) (int64, error)
}

type forumService struct {
	forumRepo       repository.ForumRepository
	forumMemberRepo repository.ForumMemberRepository
	routeRepo       repository.RouteRepository
}

func NewForumService(
	forumRepo repository.ForumRepository,
	forumMemberRepo repository.ForumMemberRepository,
	routeRepo repository.RouteRepository,
) ForumService {
	return &forumService{
		forumRepo:       forumRepo,
		forumMemberRepo: forumMemberRepo,
		routeRepo:       routeRepo,
	}
}

func (s *forumService) GetOrCreateForumByRouteID(routeID uint) (*models.Forum, error) {
	// Verify route exists
	_, err := s.routeRepo.FindByID(routeID)
	if err != nil {
		return nil, errors.New("route not found")
	}

	return s.forumRepo.FindOrCreateByRouteID(routeID)
}

func (s *forumService) GetForumByRouteID(routeID uint) (*models.Forum, error) {
	forum, err := s.forumRepo.FindByRouteID(routeID)
	if err != nil {
		return nil, errors.New("forum not found")
	}
	return forum, nil
}

func (s *forumService) GetForumByID(forumID uint) (*models.Forum, error) {
	forum, err := s.forumRepo.FindByID(forumID)
	if err != nil {
		return nil, errors.New("forum not found")
	}
	return forum, nil
}

func (s *forumService) JoinForum(forumID uint, userID uint) error {
	// Verify forum exists
	_, err := s.forumRepo.FindByID(forumID)
	if err != nil {
		return errors.New("forum not found")
	}

	// Check if already a member
	isMember, err := s.forumMemberRepo.IsMember(forumID, userID)
	if err != nil {
		return err
	}
	if isMember {
		return errors.New("already a member of this forum")
	}

	member := &models.ForumMember{
		ForumID: forumID,
		UserID:  userID,
	}

	return s.forumMemberRepo.Create(member)
}

func (s *forumService) LeaveForum(forumID uint, userID uint) error {
	// Check if member
	isMember, err := s.forumMemberRepo.IsMember(forumID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("not a member of this forum")
	}

	return s.forumMemberRepo.Delete(forumID, userID)
}

func (s *forumService) IsMember(forumID uint, userID uint) (bool, error) {
	return s.forumMemberRepo.IsMember(forumID, userID)
}

func (s *forumService) GetForumMembers(forumID uint, offset, limit int) ([]models.ForumMember, int64, error) {
	return s.forumMemberRepo.ListMembers(forumID, offset, limit)
}

func (s *forumService) GetMemberCount(forumID uint) (int64, error) {
	return s.forumMemberRepo.CountByForumID(forumID)
}

