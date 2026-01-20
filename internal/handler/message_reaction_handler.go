package handler

import (
	"net/http"
	"strconv"

	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type MessageReactionHandler struct {
	messageReactionService service.MessageReactionService
}

func NewMessageReactionHandler(messageReactionService service.MessageReactionService) *MessageReactionHandler {
	return &MessageReactionHandler{
		messageReactionService: messageReactionService,
	}
}

type CreateReactionRequest struct {
	Emoji string `json:"emoji" binding:"required"`
}

func (h *MessageReactionHandler) CreateReaction(c *gin.Context) {
	idStr := c.Param("id")
	messageID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")
	currentUserID := userID.(uint)

	var req CreateReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	reaction, err := h.messageReactionService.CreateReaction(uint(messageID), currentUserID, req.Emoji)
	if err != nil {
		BadRequest(c, err)
		return
	}

	if reaction == nil {
		MessageResponse(c, http.StatusOK, "Reaction removed (emoji toggled)")
		return
	}

	SuccessResponse(c, http.StatusCreated, reaction)
}

func (h *MessageReactionHandler) RemoveReaction(c *gin.Context) {
	idStr := c.Param("id")
	messageID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")
	currentUserID := userID.(uint)

	if err := h.messageReactionService.RemoveReaction(uint(messageID), currentUserID); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Reaction removed successfully")
}

func (h *MessageReactionHandler) GetMessageReactions(c *gin.Context) {
	idStr := c.Param("id")
	messageID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	reactions, err := h.messageReactionService.GetMessageReactions(uint(messageID))
	if err != nil {
		BadRequest(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"reactions": reactions,
	})
}
