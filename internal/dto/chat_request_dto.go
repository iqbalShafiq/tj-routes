package dto

import (
	"time"

	"tj-routes/internal/models"
)

type ChatRequestResponse struct {
	ID         uint          `json:"id"`
	SenderID   uint          `json:"sender_id"`
	ReceiverID uint          `json:"receiver_id"`
	Message    string        `json:"message"`
	Status     string        `json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	Sender     *UserBasicDTO `json:"sender,omitempty"`
	Receiver   *UserBasicDTO `json:"receiver,omitempty"`
}

type ChatRequestListResponse struct {
	Items []ChatRequestResponse `json:"items"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}

type CreateChatRequestRequest struct {
	ReceiverID uint   `json:"receiver_id" binding:"required"`
	Message    string `json:"message" binding:"required"`
}

type UpdateChatRequestStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=accepted rejected"`
}

func ToChatRequestResponse(request *models.ChatRequest) ChatRequestResponse {
	return ChatRequestResponse{
		ID:         request.ID,
		SenderID:   request.SenderID,
		ReceiverID: request.ReceiverID,
		Message:    request.Message,
		Status:     request.Status,
		CreatedAt:  request.CreatedAt,
		UpdatedAt:  request.UpdatedAt,
	}
}

func ToChatRequestResponseWithUsers(request *models.ChatRequest) ChatRequestResponse {
	response := ToChatRequestResponse(request)
	if request.Sender.ID != 0 {
		response.Sender = &UserBasicDTO{
			ID:       request.Sender.ID,
			Username: request.Sender.Username,
			Email:    request.Sender.Email,
		}
	}
	if request.Receiver.ID != 0 {
		response.Receiver = &UserBasicDTO{
			ID:       request.Receiver.ID,
			Username: request.Receiver.Username,
			Email:    request.Receiver.Email,
		}
	}
	return response
}
