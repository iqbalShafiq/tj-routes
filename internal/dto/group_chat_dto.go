package dto

import (
	"time"

	"tj-routes/internal/models"
)

type GroupChatResponse struct {
	ID          uint             `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	AvatarURL   *string          `json:"avatar_url"`
	Type        string           `json:"type"`
	MaxMembers  int              `json:"max_members"`
	CreatedByID uint             `json:"created_by_id"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	CreatedBy   *UserBasicDTO    `json:"created_by,omitempty"`
	Members     []GroupMemberDTO `json:"members,omitempty"`
}

type GroupChatListResponse struct {
	Items []GroupChatResponse `json:"items"`
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Limit int                 `json:"limit"`
}

type CreateGroupChatRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"omitempty,oneof=public private restricted"`
	MaxMembers  int    `json:"max_members"`
}

type UpdateGroupChatRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	AvatarURL   *string `json:"avatar_url"`
	Type        *string `json:"type"`
	MaxMembers  *int    `json:"max_members"`
}

func ToGroupChatResponse(group *models.GroupChat) GroupChatResponse {
	return GroupChatResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		AvatarURL:   group.AvatarURL,
		Type:        group.Type,
		MaxMembers:  group.MaxMembers,
		CreatedByID: group.CreatedByID,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}
}

func ToGroupChatResponseWithMembers(group *models.GroupChat) GroupChatResponse {
	response := ToGroupChatResponse(group)
	if group.CreatedBy.ID != 0 {
		response.CreatedBy = &UserBasicDTO{
			ID:       group.CreatedBy.ID,
			Username: group.CreatedBy.Username,
			Email:    group.CreatedBy.Email,
		}
	}
	members := make([]GroupMemberDTO, len(group.Members))
	for i, m := range group.Members {
		members[i] = GroupMemberDTO{
			ID:         m.ID,
			GroupID:    m.GroupID,
			UserID:     m.UserID,
			Role:       m.Role,
			LastReadAt: m.LastReadAt,
			MutedUntil: m.MutedUntil,
			JoinedAt:   m.JoinedAt,
		}
		if m.User.ID != 0 {
			members[i].User = &UserBasicDTO{
				ID:       m.User.ID,
				Username: m.User.Username,
				Email:    m.User.Email,
			}
		}
	}
	response.Members = members
	return response
}
