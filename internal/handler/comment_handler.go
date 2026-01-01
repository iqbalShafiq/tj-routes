package handler

import (
	"net/http"
	"strconv"

	"tj-routes/internal/models"
	"tj-routes/internal/repository"
	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	commentService service.CommentService
	reactionRepo   repository.ReactionRepository
}

type CreateCommentRequest struct {
	Content  string `json:"content" binding:"required"`
	ParentID *uint  `json:"parent_id,omitempty"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

func NewCommentHandler(commentService service.CommentService, reactionRepo repository.ReactionRepository) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
		reactionRepo:   reactionRepo,
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

	// Enhance comments with user_reaction
	enhancedComments := h.enhanceComments(c, comments)

	SuccessResponse(c, http.StatusOK, gin.H{
		"comments": enhancedComments,
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

	reportIDUint := uint(reportID)
	comment := &models.Comment{
		ReportID: &reportIDUint,
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

func (h *CommentHandler) GetCommentsByForumPost(c *gin.Context) {
	forumPostID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	comments, err := h.commentService.GetCommentsByForumPostID(uint(forumPostID))
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"comments": comments,
	})
}

func (h *CommentHandler) CreateCommentOnForumPost(c *gin.Context) {
	forumPostID, err := strconv.ParseUint(c.Param("id"), 10, 32)
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
		ForumPostID: func() *uint { id := uint(forumPostID); return &id }(),
		UserID:      userID.(uint),
		Content:    req.Content,
		ParentID:    req.ParentID,
	}

	if err := h.commentService.CreateComment(comment); err != nil {
		BadRequest(c, err)
		return
	}

	// Reload comment with user data
	comment, _ = h.commentService.GetCommentByID(comment.ID)

	SuccessResponse(c, http.StatusCreated, comment)
}

// enhanceComments adds user_reaction to comments for authenticated users
func (h *CommentHandler) enhanceComments(c *gin.Context, comments []models.Comment) []map[string]interface{} {
	var userID *uint
	if userIDVal, exists := c.Get("user_id"); exists {
		id := userIDVal.(uint)
		userID = &id
	}

	enhanced := make([]map[string]interface{}, len(comments))
	for i, comment := range comments {
		enhanced[i] = h.enhanceCommentWithReplies(comment, userID)
	}

	return enhanced
}

// enhanceCommentWithReplies adds user_reaction to a comment and recursively processes replies
func (h *CommentHandler) enhanceCommentWithReplies(comment models.Comment, userID *uint) map[string]interface{} {
	enhancedComment := map[string]interface{}{
		"id":         comment.ID,
		"report_id":  comment.ReportID,
		"user_id":    comment.UserID,
		"parent_id":  comment.ParentID,
		"content":    comment.Content,
		"upvotes":    comment.Upvotes,
		"downvotes":  comment.Downvotes,
		"created_at": comment.CreatedAt,
		"updated_at": comment.UpdatedAt,
		"user":       comment.User,
	}

	// Add user reaction if user is authenticated and reaction repo is available
	if userID != nil && h.reactionRepo != nil {
		if reaction, err := h.reactionRepo.FindByUserAndTarget(*userID, models.ReactionTargetComment, comment.ID); err == nil && reaction != nil {
			enhancedComment["user_reaction"] = reaction.ReactionType
		} else {
			enhancedComment["user_reaction"] = nil
		}
	}

	// Recursively process replies
	if len(comment.Replies) > 0 {
		enhancedReplies := make([]map[string]interface{}, len(comment.Replies))
		for i, reply := range comment.Replies {
			enhancedReplies[i] = h.enhanceCommentWithReplies(reply, userID)
		}
		enhancedComment["replies"] = enhancedReplies
	} else {
		enhancedComment["replies"] = []map[string]interface{}{}
	}

	return enhancedComment
}

