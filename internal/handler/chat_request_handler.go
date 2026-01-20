package handler

import (
	"net/http"
	"strconv"

	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type ChatRequestHandler struct {
	chatRequestService service.ChatRequestService
}

func NewChatRequestHandler(chatRequestService service.ChatRequestService) *ChatRequestHandler {
	return &ChatRequestHandler{
		chatRequestService: chatRequestService,
	}
}

type CreateChatRequestRequest struct {
	ReceiverID uint   `json:"receiver_id" binding:"required"`
	Message    string `json:"message" binding:"required"`
}

func (h *ChatRequestHandler) CreateChatRequest(c *gin.Context) {
	userID, _ := c.Get("user_id")
	senderID := userID.(uint)

	var req CreateChatRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	chatRequest, err := h.chatRequestService.CreateChatRequest(senderID, req.ReceiverID, req.Message)
	if err != nil {
		BadRequest(c, err)
		return
	}

	SuccessResponse(c, http.StatusCreated, chatRequest)
}

func (h *ChatRequestHandler) ListSentRequests(c *gin.Context) {
	userID, _ := c.Get("user_id")
	currentUserID := userID.(uint)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	requests, total, err := h.chatRequestService.ListSentRequests(currentUserID, offset, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"requests": requests,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

func (h *ChatRequestHandler) ListReceivedRequests(c *gin.Context) {
	userID, _ := c.Get("user_id")
	currentUserID := userID.(uint)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	requests, total, err := h.chatRequestService.ListReceivedRequests(currentUserID, offset, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"requests": requests,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

func (h *ChatRequestHandler) AcceptRequest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")
	currentUserID := userID.(uint)

	if err := h.chatRequestService.AcceptRequest(uint(id), currentUserID); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Chat request accepted")
}

func (h *ChatRequestHandler) RejectRequest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")
	currentUserID := userID.(uint)

	if err := h.chatRequestService.RejectRequest(uint(id), currentUserID); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Chat request rejected")
}
