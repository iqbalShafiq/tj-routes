package handler

import (
	"net/http"
	"strconv"

	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	messageService service.MessageService
}

func NewMessageHandler(messageService service.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

type CreateMessageRequest struct {
	Content        string `json:"content" binding:"required"`
	ConversationID *uint  `json:"conversation_id,omitempty"`
	GroupID        *uint  `json:"group_id,omitempty"`
	ReplyToID      *uint  `json:"reply_to_id,omitempty"`
}

func (h *MessageHandler) CreateMessage(c *gin.Context) {
	userID, _ := c.Get("user_id")
	senderID := userID.(uint)

	var req CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	message, err := h.messageService.CreateMessage(req.Content, senderID, req.ConversationID, req.GroupID)
	if err != nil {
		BadRequest(c, err)
		return
	}

	SuccessResponse(c, http.StatusCreated, message)
}

func (h *MessageHandler) GetMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	message, err := h.messageService.GetMessageByID(uint(id))
	if err != nil {
		NotFound(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, message)
}

func (h *MessageHandler) ListConversationMessages(c *gin.Context) {
	idStr := c.Param("id")
	conversationID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset := (page - 1) * limit

	messages, total, err := h.messageService.ListConversationMessages(uint(conversationID), offset, limit)
	if err != nil {
		BadRequest(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"messages": messages,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

func (h *MessageHandler) ListGroupMessages(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset := (page - 1) * limit

	messages, total, err := h.messageService.ListGroupMessages(uint(groupID), offset, limit)
	if err != nil {
		BadRequest(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"messages": messages,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")
	currentUserID := userID.(uint)

	if err := h.messageService.DeleteMessage(uint(id), currentUserID); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Message deleted successfully")
}

type UpdateMessageStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *MessageHandler) UpdateMessageStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")
	currentUserID := userID.(uint)

	var req UpdateMessageStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	if err := h.messageService.UpdateMessageStatus(uint(id), currentUserID, req.Status); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Message status updated")
}
