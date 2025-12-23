package handler

import (
	"net/http"
	"strconv"

	"tj-routes/internal/models"
	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	commentService service.CommentService
}

type CreateCommentRequest struct {
	Content  string `json:"content" binding:"required"`
	ParentID *uint  `json:"parent_id,omitempty"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

func NewCommentHandler(commentService service.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

func (h *CommentHandler) GetComments(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	comments, err := h.commentService.GetCommentsByReportID(uint(reportID))
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"comments": comments,
	})
}

func (h *CommentHandler) CreateComment(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, err)
		return
	}

	comment := &models.Comment{
		ReportID: uint(reportID),
		UserID:   userID.(uint),
		Content:  req.Content,
		ParentID: req.ParentID,
	}

	if err := h.commentService.CreateComment(comment); err != nil {
		BadRequest(c, err)
		return
	}

	// Reload comment with user data
	comment, _ = h.commentService.GetCommentByID(comment.ID)

	SuccessResponse(c, http.StatusCreated, comment)
}

func (h *CommentHandler) UpdateComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	var req UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, err)
		return
	}

	if err := h.commentService.UpdateComment(uint(id), userID.(uint), req.Content); err != nil {
		BadRequest(c, err)
		return
	}

	comment, _ := h.commentService.GetCommentByID(uint(id))
	SuccessResponse(c, http.StatusOK, comment)
}

func (h *CommentHandler) DeleteComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, err)
		return
	}

	if err := h.commentService.DeleteComment(uint(id), userID.(uint)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Comment deleted successfully")
}

