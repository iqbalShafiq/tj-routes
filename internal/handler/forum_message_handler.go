package handler

import (
	"net/http"
	"strconv"

	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type ForumMessageHandler struct {
	forumMessageService service.ForumMessageService
}

func NewForumMessageHandler(forumMessageService service.ForumMessageService) *ForumMessageHandler {
	return &ForumMessageHandler{
		forumMessageService: forumMessageService,
	}
}

type CreateForumMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

func (h *ForumMessageHandler) ListMessages(c *gin.Context) {
	forumIDStr := c.Param("id")
	forumID, err := strconv.ParseUint(forumIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset := (page - 1) * limit

	messages, total, err := h.forumMessageService.ListForumMessages(uint(forumID), offset, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"messages": messages,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

func (h *ForumMessageHandler) CreateMessage(c *gin.Context) {
	userID, _ := c.Get("user_id")
	forumIDStr := c.Param("id")
	forumID, err := strconv.ParseUint(forumIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	var req CreateForumMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	message, err := h.forumMessageService.CreateMessage(uint(forumID), userID.(uint), req.Content)
	if err != nil {
		BadRequest(c, err)
		return
	}

	SuccessResponse(c, http.StatusCreated, message)
}

func (h *ForumMessageHandler) DeleteMessage(c *gin.Context) {
	userID, _ := c.Get("user_id")
	forumIDStr := c.Param("id")
	_, err := strconv.ParseUint(forumIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	msgIDStr := c.Param("msg_id")
	msgID, err := strconv.ParseUint(msgIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	if err := h.forumMessageService.DeleteMessage(uint(msgID), userID.(uint)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Message deleted successfully")
}
