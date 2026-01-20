package service

import (
	"context"
	"errors"
	"time"

	"tj-routes/internal/cache"
	"tj-routes/internal/config"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type GroupChatService interface {
	CreateGroup(name, description, groupType string, createdBy uint, avatarURL *string) (*models.GroupChat, error)
	GetGroupByID(id uint) (*models.GroupChat, error)
	ListUserGroups(userID uint, offset, limit int) ([]models.GroupChat, int64, error)
	UpdateGroup(id uint, name, description, groupType string, avatarURL *string, requestingUserID uint) error
	DeleteGroup(id uint, requestingUserID uint) error
	UpdateGroupAvatar(id uint, avatarURL string) error
	ValidateUserCanModify(groupID, userID uint) (bool, error)
}

type groupChatService struct {
	groupRepo  repository.GroupChatRepository
	memberRepo repository.GroupMemberRepository
	userRepo   repository.UserRepository
	cache      cache.Cache
	cfg        *config.Config
}

func NewGroupChatService(
	groupRepo repository.GroupChatRepository,
	memberRepo repository.GroupMemberRepository,
	userRepo repository.UserRepository,
	cache cache.Cache,
	cfg *config.Config,
) GroupChatService {
	return &groupChatService{
		groupRepo:  groupRepo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
		cache:      cache,
		cfg:        cfg,
	}
}

func (s *groupChatService) CreateGroup(name, description, groupType string, createdBy uint, avatarURL *string) (*models.GroupChat, error) {
	if name == "" {
		return nil, errors.New("group name is required")
	}

	validTypes := map[string]bool{"public": true, "private": true, "restricted": true}
	if !validTypes[groupType] {
		groupType = "private"
	}

	if err := s.validateUserGroupLimit(createdBy); err != nil {
		return nil, err
	}

	group := &models.GroupChat{
		Name:        name,
		Description: description,
		Type:        groupType,
		AvatarURL:   avatarURL,
		MaxMembers:  s.cfg.Chat.MaxGroupMembers,
		CreatedByID: createdBy,
	}

	if err := s.groupRepo.Create(group); err != nil {
		return nil, err
	}

	member := &models.GroupMember{
		GroupID:  group.ID,
		UserID:   createdBy,
		Role:     "admin",
		JoinedAt: time.Now(),
	}

	if err := s.memberRepo.Create(member); err != nil {
		return nil, err
	}

	if s.cache != nil {
		s.invalidateGroupCache(group.ID)
		s.invalidateUserGroupsCache(createdBy)
	}

	return group, nil
}

func (s *groupChatService) GetGroupByID(id uint) (*models.GroupChat, error) {
	var group models.GroupChat
	ctx := context.Background()

	if s.cache != nil {
		cacheKey := cache.GroupChatKey(id)
		if err := s.cache.Get(ctx, cacheKey, &group); err == nil {
			return &group, nil
		}
	}

	groupData, err := s.groupRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("group not found")
	}

	if s.cache != nil {
		cacheKey := cache.GroupChatKey(id)
		s.cache.Set(ctx, cacheKey, groupData, cache.DefaultTTL)
	}

	return groupData, nil
}

func (s *groupChatService) ListUserGroups(userID uint, offset, limit int) ([]models.GroupChat, int64, error) {
	groups, total, err := s.groupRepo.ListByUserID(userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	if s.cache != nil {
		ctx := context.Background()
		cacheKey := cache.GroupListKey(userID, offset/limit+1, limit)
		s.cache.Set(ctx, cacheKey, groups, cache.DefaultTTL)
	}

	return groups, total, nil
}

func (s *groupChatService) UpdateGroup(id uint, name, description, groupType string, avatarURL *string, requestingUserID uint) error {
	group, err := s.groupRepo.FindByID(id)
	if err != nil {
		return errors.New("group not found")
	}

	canModify, err := s.ValidateUserCanModify(id, requestingUserID)
	if err != nil {
		return err
	}
	if !canModify {
		return errors.New("only admins and moderators can update group")
	}

	if name != "" {
		group.Name = name
	}
	if description != "" {
		group.Description = description
	}
	if groupType != "" {
		validTypes := map[string]bool{"public": true, "private": true, "restricted": true}
		if validTypes[groupType] {
			group.Type = groupType
		}
	}
	if avatarURL != nil {
		group.AvatarURL = avatarURL
	}

	if err := s.groupRepo.Update(group); err != nil {
		return err
	}

	if s.cache != nil {
		s.invalidateGroupCache(id)
		s.invalidateUserGroupsCache(requestingUserID)
	}

	return nil
}

func (s *groupChatService) DeleteGroup(id uint, requestingUserID uint) error {
	group, err := s.groupRepo.FindByID(id)
	if err != nil {
		return errors.New("group not found")
	}

	if group.CreatedByID != requestingUserID {
		return errors.New("only the group creator can delete the group")
	}

	if err := s.groupRepo.Delete(id); err != nil {
		return err
	}

	if s.cache != nil {
		s.invalidateGroupCache(id)
		s.invalidateUserGroupsCache(requestingUserID)
	}

	return nil
}

func (s *groupChatService) UpdateGroupAvatar(id uint, avatarURL string) error {
	if err := s.groupRepo.UpdateAvatar(id, avatarURL); err != nil {
		return err
	}

	if s.cache != nil {
		s.invalidateGroupCache(id)
	}

	return nil
}

func (s *groupChatService) ValidateUserCanModify(groupID, userID uint) (bool, error) {
	member, err := s.memberRepo.FindByGroupAndUser(groupID, userID)
	if err != nil {
		return false, nil
	}

	return member.Role == "admin" || member.Role == "moderator", nil
}

func (s *groupChatService) validateUserGroupLimit(userID uint) error {
	_, total, err := s.memberRepo.ListByUserID(userID, 0, s.cfg.Chat.MaxGroupMembersPerUser+1)
	if err != nil {
		return err
	}

	if total >= int64(s.cfg.Chat.MaxGroupMembersPerUser) {
		return errors.New("user has reached maximum group limit")
	}

	return nil
}

func (s *groupChatService) invalidateGroupCache(id uint) {
	if s.cache == nil {
		return
	}
	ctx := context.Background()
	cacheKey := cache.GroupChatKey(id)
	s.cache.Delete(ctx, cacheKey)
}

func (s *groupChatService) invalidateUserGroupsCache(userID uint) {
	if s.cache == nil {
		return
	}
	ctx := context.Background()
	pattern := "chat:user:" + string(rune(userID)) + ":groups:*"
	s.cache.InvalidatePattern(ctx, pattern)
}
