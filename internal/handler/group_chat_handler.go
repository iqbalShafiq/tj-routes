package handler

import (
	"net/http"
	"strconv"

	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type GroupChatHandler struct {
	groupChatService service.GroupChatService
	memberService    service.GroupMemberService
}

func NewGroupChatHandler(groupChatService service.GroupChatService, memberService service.GroupMemberService) *GroupChatHandler {
	return &GroupChatHandler{
		groupChatService: groupChatService,
		memberService:    memberService,
	}
}

func (h *GroupChatHandler) CreateGroup(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, nil)
		return
	}

	group, err := h.groupChatService.CreateGroup(req.Name, req.Description, req.Type, userID.(uint), nil)
	if err != nil {
		BadRequest(c, err)
		return
	}

	SuccessResponse(c, http.StatusCreated, group)
}

func (h *GroupChatHandler) GetGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	group, err := h.groupChatService.GetGroupByID(uint(id))
	if err != nil {
		NotFound(c, err)
		return
	}

	memberCount, _ := h.memberService.GetMemberCount(group.ID)

	var isMember bool
	var role string
	if userID, exists := c.Get("user_id"); exists {
		isMember, _ = h.memberService.IsMember(group.ID, userID.(uint))
		role, _ = h.memberService.GetMemberRole(group.ID, userID.(uint))
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"group":        group,
		"member_count": memberCount,
		"is_member":    isMember,
		"role":         role,
	})
}

func (h *GroupChatHandler) ListUserGroups(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	groups, total, err := h.groupChatService.ListUserGroups(userID.(uint), offset, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"items": groups,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *GroupChatHandler) UpdateGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")

	if err := h.groupChatService.UpdateGroup(uint(id), req.Name, req.Description, req.Type, nil, userID.(uint)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Group updated successfully")
}

func (h *GroupChatHandler) DeleteGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")

	if err := h.groupChatService.DeleteGroup(uint(id), userID.(uint)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Group deleted successfully")
}

func (h *GroupChatHandler) UpdateAvatar(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	var req struct {
		AvatarURL string `json:"avatar_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")

	canModify, _ := h.groupChatService.ValidateUserCanModify(uint(id), userID.(uint))
	if !canModify {
		Forbidden(c, nil)
		return
	}

	if err := h.groupChatService.UpdateGroupAvatar(uint(id), req.AvatarURL); err != nil {
		InternalServerError(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Avatar updated successfully")
}
