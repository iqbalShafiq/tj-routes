package dto

import (
	"time"

	"tj-routes/internal/models"
)

type GroupInviteResponse struct {
	ID        uint               `json:"id"`
	GroupID   uint               `json:"group_id"`
	InviterID uint               `json:"inviter_id"`
	InviteeID uint               `json:"invitee_id"`
	Status    string             `json:"status"`
	ExpiresAt time.Time          `json:"expires_at"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	Group     *GroupChatResponse `json:"group,omitempty"`
	Inviter   *UserBasicDTO      `json:"inviter,omitempty"`
	Invitee   *UserBasicDTO      `json:"invitee,omitempty"`
}

type GroupInviteListResponse struct {
	Items []GroupInviteResponse `json:"items"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}

type CreateGroupInviteRequest struct {
	InviteeID uint `json:"invitee_id" binding:"required"`
}

type RespondGroupInviteRequest struct {
	Status string `json:"status" binding:"required,oneof=accepted rejected"`
}

func ToGroupInviteResponse(invite *models.GroupInvite) GroupInviteResponse {
	return GroupInviteResponse{
		ID:        invite.ID,
		GroupID:   invite.GroupID,
		InviterID: invite.InviterID,
		InviteeID: invite.InviteeID,
		Status:    invite.Status,
		ExpiresAt: invite.ExpiresAt,
		CreatedAt: invite.CreatedAt,
		UpdatedAt: invite.UpdatedAt,
	}
}

func ToGroupInviteResponseWithUsers(invite *models.GroupInvite) GroupInviteResponse {
	response := ToGroupInviteResponse(invite)
	if invite.Group.ID != 0 {
		groupResp := ToGroupChatResponse(&invite.Group)
		response.Group = &groupResp
	}
	if invite.Inviter.ID != 0 {
		response.Inviter = &UserBasicDTO{
			ID:       invite.Inviter.ID,
			Username: invite.Inviter.Username,
			Email:    invite.Inviter.Email,
		}
	}
	if invite.Invitee.ID != 0 {
		response.Invitee = &UserBasicDTO{
			ID:       invite.Invitee.ID,
			Username: invite.Invitee.Username,
			Email:    invite.Invitee.Email,
		}
	}
	return response
}
