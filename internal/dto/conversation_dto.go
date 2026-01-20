package dto

import (
	"time"

	"tj-routes/internal/models"
)

type ConversationResponse struct {
	ID           uint                         `json:"id"`
	Type         string                       `json:"type"`
	CreatedAt    time.Time                    `json:"created_at"`
	UpdatedAt    time.Time                    `json:"updated_at"`
	Participants []ConversationParticipantDTO `json:"participants,omitempty"`
}

type ConversationListResponse struct {
	Items []ConversationResponse `json:"items"`
	Total int64                  `json:"total"`
	Page  int                    `json:"page"`
	Limit int                    `json:"limit"`
}

type ConversationParticipantDTO struct {
	ID             uint          `json:"id"`
	ConversationID uint          `json:"conversation_id"`
	UserID         uint          `json:"user_id"`
	LastReadAt     *time.Time    `json:"last_read_at"`
	JoinedAt       time.Time     `json:"joined_at"`
	User           *UserBasicDTO `json:"user,omitempty"`
}

type CreateConversationRequest struct {
	Type           string `json:"type"`
	ParticipantIDs []uint `json:"participant_ids"`
	Message        string `json:"message,omitempty"`
}

func ToConversationResponse(conversation *models.Conversation) ConversationResponse {
	return ConversationResponse{
		ID:        conversation.ID,
		Type:      conversation.Type,
		CreatedAt: conversation.CreatedAt,
		UpdatedAt: conversation.UpdatedAt,
	}
}

func ToConversationResponseWithParticipants(conversation *models.Conversation) ConversationResponse {
	response := ToConversationResponse(conversation)
	participants := make([]ConversationParticipantDTO, len(conversation.Participants))
	for i, p := range conversation.Participants {
		participants[i] = ConversationParticipantDTO{
			ID:             p.ID,
			ConversationID: p.ConversationID,
			UserID:         p.UserID,
			LastReadAt:     p.LastReadAt,
			JoinedAt:       p.JoinedAt,
		}
		if p.User.ID != 0 {
			participants[i].User = &UserBasicDTO{
				ID:       p.User.ID,
				Username: p.User.Username,
				Email:    p.User.Email,
			}
		}
	}
	response.Participants = participants
	return response
}
