package dto

import (
	"time"

	"tj-routes/internal/models"
)

type GroupMemberDTO struct {
	ID         uint               `json:"id"`
	GroupID    uint               `json:"group_id"`
	UserID     uint               `json:"user_id"`
	Role       string             `json:"role"`
	LastReadAt *time.Time         `json:"last_read_at"`
	MutedUntil *time.Time         `json:"muted_until"`
	JoinedAt   time.Time          `json:"joined_at"`
	Group      *GroupChatResponse `json:"group,omitempty"`
	User       *UserBasicDTO      `json:"user,omitempty"`
}

type GroupMemberListResponse struct {
	Items []GroupMemberDTO `json:"items"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

type AddGroupMemberRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"omitempty,oneof=admin moderator member"`
}

type UpdateGroupMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin moderator member"`
}

type UpdateGroupMemberMuteRequest struct {
	MutedUntil *time.Time `json:"muted_until"`
}

func ToGroupMemberDTO(member *models.GroupMember) GroupMemberDTO {
	return GroupMemberDTO{
		ID:         member.ID,
		GroupID:    member.GroupID,
		UserID:     member.UserID,
		Role:       member.Role,
		LastReadAt: member.LastReadAt,
		MutedUntil: member.MutedUntil,
		JoinedAt:   member.JoinedAt,
	}
}

func ToGroupMemberDTOWithUser(member *models.GroupMember) GroupMemberDTO {
	response := ToGroupMemberDTO(member)
	if member.User.ID != 0 {
		response.User = &UserBasicDTO{
			ID:       member.User.ID,
			Username: member.User.Username,
			Email:    member.User.Email,
		}
	}
	return response
}
