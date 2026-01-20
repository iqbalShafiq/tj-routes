package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"tj-routes/internal/cache"
	"tj-routes/internal/config"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type GroupMemberService interface {
	AddMember(groupID, userID uint, role string, addedBy uint) error
	RemoveMember(groupID, userID, requestingUserID uint) error
	ListGroupMembers(groupID uint, offset, limit int) ([]models.GroupMember, int64, error)
	UpdateMemberRole(groupID, userID uint, newRole string, requestingUserID uint) error
	UpdateLastRead(groupID, userID uint) error
	UpdateMutedUntil(groupID, userID uint, mutedUntil *time.Time) error
	GetMemberCount(groupID uint) (int64, error)
	IsMember(groupID, userID uint) (bool, error)
	GetMemberRole(groupID, userID uint) (string, error)
}

type groupMemberService struct {
	groupRepo    repository.GroupChatRepository
	memberRepo   repository.GroupMemberRepository
	groupService GroupChatService
	cache        cache.Cache
	cfg          *config.Config
}

func NewGroupMemberService(
	groupRepo repository.GroupChatRepository,
	memberRepo repository.GroupMemberRepository,
	groupService GroupChatService,
	cache cache.Cache,
	cfg *config.Config,
) GroupMemberService {
	return &groupMemberService{
		groupRepo:    groupRepo,
		memberRepo:   memberRepo,
		groupService: groupService,
		cache:        cache,
		cfg:          cfg,
	}
}

func (s *groupMemberService) AddMember(groupID, userID uint, role string, addedBy uint) error {
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return errors.New("group not found")
	}

	canModify, _ := s.groupService.ValidateUserCanModify(groupID, addedBy)
	if !canModify {
		return errors.New("only admins can add members")
	}

	isMember, err := s.IsMember(groupID, userID)
	if err != nil {
		return err
	}
	if isMember {
		return errors.New("user is already a member")
	}

	memberCount, _ := s.GetMemberCount(groupID)
	if memberCount >= int64(group.MaxMembers) {
		return errors.New("group is full")
	}

	_, err = s.memberRepo.FindByGroupAndUser(groupID, addedBy)
	if err != nil {
		return errors.New("requesting user is not a member of this group")
	}

	validRoles := map[string]bool{"admin": true, "moderator": true, "member": true}
	if !validRoles[role] {
		role = "member"
	}

	member := &models.GroupMember{
		GroupID:  groupID,
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now(),
	}

	if err := s.memberRepo.Create(member); err != nil {
		return err
	}

	s.invalidateGroupMembersCache(groupID)
	s.invalidateUserGroupsCache(userID)

	return nil
}

func (s *groupMemberService) RemoveMember(groupID, userID, requestingUserID uint) error {
	isMember, _ := s.IsMember(groupID, requestingUserID)
	if !isMember {
		return errors.New("requesting user is not a member of this group")
	}

	if userID == requestingUserID {
		return s.memberRepo.Delete(groupID, userID)
	}

	canModify, _ := s.groupService.ValidateUserCanModify(groupID, requestingUserID)
	if !canModify {
		return errors.New("only admins and moderators can remove members")
	}

	targetMember, err := s.memberRepo.FindByGroupAndUser(groupID, requestingUserID)
	if err != nil {
		return errors.New("requesting user is not a member of this group")
	}

	if targetMember.Role == "admin" && requestingUserID != userID {
		requestingMember, _ := s.memberRepo.FindByGroupAndUser(groupID, requestingUserID)
		if requestingMember.Role != "admin" {
			return errors.New("cannot remove other admins")
		}
	}

	if err := s.memberRepo.Delete(groupID, userID); err != nil {
		return err
	}

	s.invalidateGroupMembersCache(groupID)
	s.invalidateUserGroupsCache(userID)

	return nil
}

func (s *groupMemberService) ListGroupMembers(groupID uint, offset, limit int) ([]models.GroupMember, int64, error) {
	return s.memberRepo.ListByGroupID(groupID, offset, limit)
}

func (s *groupMemberService) UpdateMemberRole(groupID, userID uint, newRole string, requestingUserID uint) error {
	canModify, _ := s.groupService.ValidateUserCanModify(groupID, requestingUserID)
	if !canModify {
		return errors.New("only admins can update roles")
	}

	validRoles := map[string]bool{"admin": true, "moderator": true, "member": true}
	if !validRoles[newRole] {
		return errors.New("invalid role")
	}

	if newRole == "admin" {
		group, err := s.groupRepo.FindByID(groupID)
		if err != nil {
			return errors.New("group not found")
		}
		if group.CreatedByID != requestingUserID {
			return errors.New("only group creator can promote to admin")
		}
	}

	if err := s.memberRepo.UpdateRole(groupID, userID, newRole); err != nil {
		return err
	}

	s.invalidateGroupMembersCache(groupID)

	return nil
}

func (s *groupMemberService) UpdateLastRead(groupID, userID uint) error {
	now := time.Now()
	return s.memberRepo.UpdateLastRead(groupID, userID, now)
}

func (s *groupMemberService) UpdateMutedUntil(groupID, userID uint, mutedUntil *time.Time) error {
	return s.memberRepo.UpdateMutedUntil(groupID, userID, mutedUntil)
}

func (s *groupMemberService) GetMemberCount(groupID uint) (int64, error) {
	return s.memberRepo.CountByGroupID(groupID)
}

func (s *groupMemberService) IsMember(groupID, userID uint) (bool, error) {
	_, err := s.memberRepo.FindByGroupAndUser(groupID, userID)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (s *groupMemberService) GetMemberRole(groupID, userID uint) (string, error) {
	member, err := s.memberRepo.FindByGroupAndUser(groupID, userID)
	if err != nil {
		return "", err
	}
	return member.Role, nil
}

func (s *groupMemberService) invalidateGroupMembersCache(groupID uint) {
	if s.cache == nil {
		return
	}
	ctx := context.Background()
	cacheKey := cache.GroupMembersKey(groupID)
	s.cache.Delete(ctx, cacheKey)
}

func (s *groupMemberService) invalidateUserGroupsCache(userID uint) {
	if s.cache == nil {
		return
	}
	ctx := context.Background()
	pattern := fmt.Sprintf("chat:user:%d:groups:*", userID)
	s.cache.InvalidatePattern(ctx, pattern)
}
