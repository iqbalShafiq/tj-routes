package service

import (
	"errors"
	"time"

	"tj-routes/internal/cache"
	"tj-routes/internal/config"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type GroupInviteService interface {
	CreateInvite(groupID, inviterID, inviteeID uint) error
	ListGroupInvites(groupID uint, offset, limit int) ([]models.GroupInvite, int64, error)
	ListUserInvites(inviteeID uint, offset, limit int) ([]models.GroupInvite, int64, error)
	AcceptInvite(id, inviteeID uint) error
	RejectInvite(id, inviteeID uint) error
	RevokeInvite(id, inviterID uint) error
	MarkExpiredInvites() error
}

type groupInviteService struct {
	groupRepo     repository.GroupChatRepository
	memberRepo    repository.GroupMemberRepository
	inviteRepo    repository.GroupInviteRepository
	groupService  GroupChatService
	memberService GroupMemberService
	cache         cache.Cache
	cfg           *config.Config
}

func NewGroupInviteService(
	groupRepo repository.GroupChatRepository,
	memberRepo repository.GroupMemberRepository,
	inviteRepo repository.GroupInviteRepository,
	groupService GroupChatService,
	memberService GroupMemberService,
	cache cache.Cache,
	cfg *config.Config,
) GroupInviteService {
	return &groupInviteService{
		groupRepo:     groupRepo,
		memberRepo:    memberRepo,
		inviteRepo:    inviteRepo,
		groupService:  groupService,
		memberService: memberService,
		cache:         cache,
		cfg:           cfg,
	}
}

func (s *groupInviteService) CreateInvite(groupID, inviterID, inviteeID uint) error {
	if inviterID == inviteeID {
		return errors.New("cannot invite yourself")
	}

	canModify, _ := s.groupService.ValidateUserCanModify(groupID, inviterID)
	if !canModify {
		return errors.New("only admins can create invites")
	}

	isMember, _ := s.memberService.IsMember(groupID, inviteeID)
	if isMember {
		return errors.New("user is already a member")
	}

	existingInvite, err := s.inviteRepo.FindByGroupAndInvitee(groupID, inviteeID)
	if err == nil && existingInvite.Status == "pending" {
		return errors.New("invite already exists")
	}

	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return errors.New("group not found")
	}

	memberCount, _ := s.memberService.GetMemberCount(groupID)
	if memberCount >= int64(group.MaxMembers) {
		return errors.New("group is full")
	}

	_, err = s.memberRepo.FindByGroupAndUser(groupID, inviterID)
	if err != nil {
		return errors.New("inviter is not a member of this group")
	}

	expiresAt := time.Now().Add(time.Duration(s.cfg.Chat.InviteExpiryHours) * time.Hour)

	invite := &models.GroupInvite{
		GroupID:   groupID,
		InviterID: inviterID,
		InviteeID: inviteeID,
		Status:    "pending",
		ExpiresAt: expiresAt,
	}

	if err := s.inviteRepo.Create(invite); err != nil {
		return err
	}

	return nil
}

func (s *groupInviteService) ListGroupInvites(groupID uint, offset, limit int) ([]models.GroupInvite, int64, error) {
	invites, _, err := s.inviteRepo.ListByGroupID(groupID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	var filteredInvites []models.GroupInvite
	for _, invite := range invites {
		if invite.Status == "pending" {
			filteredInvites = append(filteredInvites, invite)
		}
	}

	return filteredInvites, int64(len(filteredInvites)), nil
}

func (s *groupInviteService) ListUserInvites(inviteeID uint, offset, limit int) ([]models.GroupInvite, int64, error) {
	invites, _, err := s.inviteRepo.ListByInviteeID(inviteeID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	var filteredInvites []models.GroupInvite
	for _, invite := range invites {
		if invite.Status == "pending" {
			filteredInvites = append(filteredInvites, invite)
		}
	}

	return filteredInvites, int64(len(filteredInvites)), nil
}

func (s *groupInviteService) AcceptInvite(id, inviteeID uint) error {
	invite, err := s.inviteRepo.FindByID(id)
	if err != nil {
		return errors.New("invite not found")
	}

	if invite.InviteeID != inviteeID {
		return errors.New("not authorized to accept this invite")
	}

	if invite.Status != "pending" {
		return errors.New("invite is not pending")
	}

	if time.Now().After(invite.ExpiresAt) {
		s.inviteRepo.UpdateStatus(id, "expired")
		return errors.New("invite has expired")
	}

	group, err := s.groupRepo.FindByID(invite.GroupID)
	if err != nil {
		return errors.New("group not found")
	}

	memberCount, _ := s.memberService.GetMemberCount(invite.GroupID)
	if memberCount >= int64(group.MaxMembers) {
		return errors.New("group is full")
	}

	isMember, _ := s.memberService.IsMember(invite.GroupID, inviteeID)
	if isMember {
		return errors.New("already a member of this group")
	}

	memberRole := "member"
	if err := s.memberService.AddMember(invite.GroupID, inviteeID, memberRole, invite.InviterID); err != nil {
		return err
	}

	if err := s.inviteRepo.UpdateStatus(id, "accepted"); err != nil {
		return err
	}

	return nil
}

func (s *groupInviteService) RejectInvite(id, inviteeID uint) error {
	invite, err := s.inviteRepo.FindByID(id)
	if err != nil {
		return errors.New("invite not found")
	}

	if invite.InviteeID != inviteeID {
		return errors.New("not authorized to reject this invite")
	}

	if invite.Status != "pending" {
		return errors.New("invite is not pending")
	}

	if err := s.inviteRepo.UpdateStatus(id, "rejected"); err != nil {
		return err
	}

	return nil
}

func (s *groupInviteService) RevokeInvite(id, inviterID uint) error {
	invite, err := s.inviteRepo.FindByID(id)
	if err != nil {
		return errors.New("invite not found")
	}

	canModify, _ := s.groupService.ValidateUserCanModify(invite.GroupID, inviterID)
	if !canModify {
		return errors.New("only admins can revoke invites")
	}

	if invite.Status != "pending" {
		return errors.New("invite is not pending")
	}

	if err := s.inviteRepo.UpdateStatus(id, "revoked"); err != nil {
		return err
	}

	return nil
}

func (s *groupInviteService) MarkExpiredInvites() error {
	return s.inviteRepo.MarkExpired()
}
