package handler

import (
	"net/http"
	"strconv"

	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type ForumHandler struct {
	forumService service.ForumService
}

func NewForumHandler(forumService service.ForumService) *ForumHandler {
	return &ForumHandler{
		forumService: forumService,
	}
}

func (h *ForumHandler) GetForumByRoute(c *gin.Context) {
	routeIDStr := c.Param("id")
	routeID, err := strconv.ParseUint(routeIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	forum, err := h.forumService.GetOrCreateForumByRouteID(uint(routeID))
	if err != nil {
		NotFound(c, err)
		return
	}

	// Get member count
	memberCount, _ := h.forumService.GetMemberCount(forum.ID)

	// Check if current user is a member (if authenticated)
	var isMember bool
	if userID, exists := c.Get("user_id"); exists {
		isMember, _ = h.forumService.IsMember(forum.ID, userID.(uint))
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"forum":       forum,
		"member_count": memberCount,
		"is_member":    isMember,
	})
}

func (h *ForumHandler) GetForum(c *gin.Context) {
	forumIDStr := c.Param("id")
	forumID, err := strconv.ParseUint(forumIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	forum, err := h.forumService.GetForumByID(uint(forumID))
	if err != nil {
		NotFound(c, err)
		return
	}

	// Get member count
	memberCount, _ := h.forumService.GetMemberCount(forum.ID)

	// Check if current user is a member (if authenticated)
	var isMember bool
	if userID, exists := c.Get("user_id"); exists {
		isMember, _ = h.forumService.IsMember(forum.ID, userID.(uint))
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"forum":       forum,
		"member_count": memberCount,
		"is_member":    isMember,
	})
}

func (h *ForumHandler) JoinForum(c *gin.Context) {
	userID, _ := c.Get("user_id")
	forumIDStr := c.Param("id")
	forumID, err := strconv.ParseUint(forumIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	if err := h.forumService.JoinForum(uint(forumID), userID.(uint)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Successfully joined forum")
}

func (h *ForumHandler) LeaveForum(c *gin.Context) {
	userID, _ := c.Get("user_id")
	forumIDStr := c.Param("id")
	forumID, err := strconv.ParseUint(forumIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	if err := h.forumService.LeaveForum(uint(forumID), userID.(uint)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Successfully left forum")
}

func (h *ForumHandler) GetForumMembers(c *gin.Context) {
	forumIDStr := c.Param("id")
	forumID, err := strconv.ParseUint(forumIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	members, total, err := h.forumService.GetForumMembers(uint(forumID), offset, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"members": members,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

func (h *ForumHandler) CheckMembership(c *gin.Context) {
	userID, _ := c.Get("user_id")
	forumIDStr := c.Param("id")
	forumID, err := strconv.ParseUint(forumIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	isMember, err := h.forumService.IsMember(uint(forumID), userID.(uint))
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"is_member": isMember,
	})
}

