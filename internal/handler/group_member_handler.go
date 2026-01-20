package handler

import (
	"net/http"
	"strconv"
	"time"

	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type GroupMemberHandler struct {
	memberService service.GroupMemberService
}

func NewGroupMemberHandler(memberService service.GroupMemberService) *GroupMemberHandler {
	return &GroupMemberHandler{
		memberService: memberService,
	}
}

func (h *GroupMemberHandler) AddMember(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	var req struct {
		UserID uint   `json:"user_id" binding:"required"`
		Role   string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")

	if err := h.memberService.AddMember(uint(groupID), req.UserID, req.Role, userID.(uint)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Member added successfully")
}

func (h *GroupMemberHandler) RemoveMember(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userIDStr := c.Param("userId")
	userIDToRemove, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")

	if err := h.memberService.RemoveMember(uint(groupID), uint(userIDToRemove), userID.(uint)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Member removed successfully")
}

func (h *GroupMemberHandler) ListMembers(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	members, total, err := h.memberService.ListGroupMembers(uint(groupID), offset, limit)
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

func (h *GroupMemberHandler) UpdateMemberRole(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userIDStr := c.Param("userId")
	userIDToUpdate, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")

	if err := h.memberService.UpdateMemberRole(uint(groupID), uint(userIDToUpdate), req.Role, userID.(uint)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Member role updated successfully")
}

func (h *GroupMemberHandler) UpdateLastRead(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")

	if err := h.memberService.UpdateLastRead(uint(groupID), userID.(uint)); err != nil {
		InternalServerError(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Last read updated successfully")
}

func (h *GroupMemberHandler) UpdateMutedUntil(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	var req struct {
		MutedUntil *time.Time `json:"muted_until"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")

	if err := h.memberService.UpdateMutedUntil(uint(groupID), userID.(uint), req.MutedUntil); err != nil {
		InternalServerError(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Mute status updated successfully")
}

func (h *GroupMemberHandler) GetMemberCount(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	count, err := h.memberService.GetMemberCount(uint(groupID))
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"count": count,
	})
}

func (h *GroupMemberHandler) CheckMembership(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, nil)
		return
	}

	isMember, err := h.memberService.IsMember(uint(groupID), userID.(uint))
	if err != nil {
		InternalServerError(c, err)
		return
	}

	role, _ := h.memberService.GetMemberRole(uint(groupID), userID.(uint))

	SuccessResponse(c, http.StatusOK, gin.H{
		"is_member": isMember,
		"role":      role,
	})
}
