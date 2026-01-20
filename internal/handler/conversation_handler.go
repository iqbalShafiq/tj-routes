package handler

import (
	"errors"
	"net/http"
	"strconv"

	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type ConversationHandler struct {
	conversationService service.ConversationService
}

func NewConversationHandler(conversationService service.ConversationService) *ConversationHandler {
	return &ConversationHandler{
		conversationService: conversationService,
	}
}

type CreateConversationRequest struct {
	UserID2 uint `json:"user_id_2" binding:"required"`
}

func (h *ConversationHandler) CreateConversation(c *gin.Context) {
	userID, _ := c.Get("user_id")
	currentUserID := userID.(uint)

	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	conversation, err := h.conversationService.GetOrCreateConversation(currentUserID, req.UserID2)
	if err != nil {
		BadRequest(c, err)
		return
	}

	SuccessResponse(c, http.StatusCreated, conversation)
}

func (h *ConversationHandler) GetConversation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	conversation, err := h.conversationService.GetConversationByID(uint(id))
	if err != nil {
		NotFound(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, conversation)
}

func (h *ConversationHandler) ListConversations(c *gin.Context) {
	userID, _ := c.Get("user_id")
	currentUserID := userID.(uint)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	conversations, total, err := h.conversationService.ListUserConversations(currentUserID, offset, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"conversations": conversations,
		"total":         total,
		"page":          page,
		"limit":         limit,
	})
}

func (h *ConversationHandler) DeleteConversation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")
	currentUserID := userID.(uint)

	if err := h.conversationService.DeleteConversation(uint(id), currentUserID); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Conversation deleted successfully")
}

func (h *ConversationHandler) GetMessages(c *gin.Context) {
	idStr := c.Param("id")
	conversationID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")
	currentUserID := userID.(uint)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset := (page - 1) * limit

	messages, total, err := h.conversationService.GetMessages(uint(conversationID), currentUserID, offset, limit)
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

func (h *ConversationHandler) MarkAsRead(c *gin.Context) {
	idStr := c.Param("id")
	conversationID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")
	currentUserID := userID.(uint)

	if err := h.conversationService.MarkAsRead(uint(conversationID), currentUserID); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Conversation marked as read")
}

func (h *ConversationHandler) AddParticipant(c *gin.Context) {
	idStr := c.Param("id")
	conversationID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	// Get requesting user ID from context
	requestingUserID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, errors.New("user not authenticated"))
		return
	}

	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	if err := h.conversationService.AddParticipant(uint(conversationID), req.UserID, requestingUserID.(uint)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Participant added successfully")
}

func (h *ConversationHandler) RemoveParticipant(c *gin.Context) {
	idStr := c.Param("id")
	conversationID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	if err := h.conversationService.RemoveParticipant(uint(conversationID), uint(userID)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Participant removed successfully")
}
