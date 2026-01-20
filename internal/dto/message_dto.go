package dto

import (
	"time"

	"tj-routes/internal/models"
)

type MessageResponse struct {
	ID             uint                 `json:"id"`
	ConversationID *uint                `json:"conversation_id,omitempty"`
	GroupID        *uint                `json:"group_id,omitempty"`
	SenderID       uint                 `json:"sender_id"`
	Content        string               `json:"content"`
	MessageType    string               `json:"message_type"`
	ReplyToID      *uint                `json:"reply_to_id,omitempty"`
	Status         string               `json:"status"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
	Sender         *UserBasicDTO        `json:"sender,omitempty"`
	ReplyTo        *MessageResponse     `json:"reply_to,omitempty"`
	Reactions      []MessageReactionDTO `json:"reactions,omitempty"`
}

type MessageListResponse struct {
	Items []MessageResponse `json:"items"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
}

type CreateMessageRequest struct {
	ConversationID *uint  `json:"conversation_id,omitempty"`
	GroupID        *uint  `json:"group_id,omitempty"`
	Content        string `json:"content" binding:"required"`
	MessageType    string `json:"message_type" binding:"omitempty,oneof=text image file system"`
	ReplyToID      *uint  `json:"reply_to_id,omitempty"`
}

type UpdateMessageRequest struct {
	Content *string `json:"content"`
}

type UpdateMessageStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=sent delivered read"`
}

type MessageReactionDTO struct {
	ID        uint          `json:"id"`
	MessageID uint          `json:"message_id"`
	UserID    uint          `json:"user_id"`
	Emoji     string        `json:"emoji"`
	CreatedAt time.Time     `json:"created_at"`
	User      *UserBasicDTO `json:"user,omitempty"`
}

type AddMessageReactionRequest struct {
	Emoji string `json:"emoji" binding:"required,max=50"`
}

func ToMessageResponse(message *models.Message) MessageResponse {
	return MessageResponse{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		GroupID:        message.GroupID,
		SenderID:       message.SenderID,
		Content:        message.Content,
		MessageType:    message.MessageType,
		ReplyToID:      message.ReplyToID,
		Status:         message.Status,
		CreatedAt:      message.CreatedAt,
		UpdatedAt:      message.UpdatedAt,
	}
}

func ToMessageResponseWithSender(message *models.Message) MessageResponse {
	response := ToMessageResponse(message)
	if message.Sender.ID != 0 {
		response.Sender = &UserBasicDTO{
			ID:       message.Sender.ID,
			Username: message.Sender.Username,
			Email:    message.Sender.Email,
		}
	}
	return response
}

func ToMessageResponseFull(message *models.Message) MessageResponse {
	response := ToMessageResponseWithSender(message)
	if message.ReplyTo != nil {
		replyResp := ToMessageResponseWithSender(message.ReplyTo)
		response.ReplyTo = &replyResp
	}
	reactions := make([]MessageReactionDTO, len(message.Reactions))
	for i, r := range message.Reactions {
		reactions[i] = MessageReactionDTO{
			ID:        r.ID,
			MessageID: r.MessageID,
			UserID:    r.UserID,
			Emoji:     r.Emoji,
			CreatedAt: r.CreatedAt,
		}
		if r.User.ID != 0 {
			reactions[i].User = &UserBasicDTO{
				ID:       r.User.ID,
				Username: r.User.Username,
				Email:    r.User.Email,
			}
		}
	}
	response.Reactions = reactions
	return response
}
