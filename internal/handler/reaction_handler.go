package handler

import (
	"net/http"
	"strconv"

	"tj-routes/internal/models"
	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type ReactionHandler struct {
	reactionService service.ReactionService
}

type ReactRequest struct {
	Type string `json:"type" binding:"required,oneof=upvote downvote"`
}

func NewReactionHandler(reactionService service.ReactionService) *ReactionHandler {
	return &ReactionHandler{
		reactionService: reactionService,
	}
}

func (h *ReactionHandler) ReactToReport(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	var req ReactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, err)
		return
	}

	var reactionType models.ReactionType
	if req.Type == "upvote" {
		reactionType = models.ReactionUpvote
	} else {
		reactionType = models.ReactionDownvote
	}

	if err := h.reactionService.ToggleReaction(userID.(uint), models.ReactionTargetReport, uint(reportID), reactionType); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Reaction updated successfully")
}

func (h *ReactionHandler) RemoveReactionFromReport(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, err)
		return
	}

	if err := h.reactionService.RemoveReaction(userID.(uint), models.ReactionTargetReport, uint(reportID)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Reaction removed successfully")
}

func (h *ReactionHandler) ReactToComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	var req ReactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, err)
		return
	}

	var reactionType models.ReactionType
	if req.Type == "upvote" {
		reactionType = models.ReactionUpvote
	} else {
		reactionType = models.ReactionDownvote
	}

	if err := h.reactionService.ToggleReaction(userID.(uint), models.ReactionTargetComment, uint(commentID), reactionType); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Reaction updated successfully")
}

func (h *ReactionHandler) RemoveReactionFromComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, err)
		return
	}

	if err := h.reactionService.RemoveReaction(userID.(uint), models.ReactionTargetComment, uint(commentID)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Reaction removed successfully")
}

