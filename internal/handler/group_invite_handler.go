package handler

import (
	"net/http"
	"strconv"

	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type GroupInviteHandler struct {
	inviteService service.GroupInviteService
}

func NewGroupInviteHandler(inviteService service.GroupInviteService) *GroupInviteHandler {
	return &GroupInviteHandler{
		inviteService: inviteService,
	}
}

func (h *GroupInviteHandler) CreateInvite(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	var req struct {
		InviteeID uint `json:"invitee_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")

	if err := h.inviteService.CreateInvite(uint(groupID), userID.(uint), req.InviteeID); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusCreated, "Invite created successfully")
}

func (h *GroupInviteHandler) ListGroupInvites(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	invites, total, err := h.inviteService.ListGroupInvites(uint(groupID), offset, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"invites": invites,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

func (h *GroupInviteHandler) ListUserInvites(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	invites, total, err := h.inviteService.ListUserInvites(userID.(uint), offset, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"invites": invites,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

func (h *GroupInviteHandler) AcceptInvite(c *gin.Context) {
	idStr := c.Param("inviteId")
	inviteID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")

	if err := h.inviteService.AcceptInvite(uint(inviteID), userID.(uint)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Invite accepted successfully")
}

func (h *GroupInviteHandler) RejectInvite(c *gin.Context) {
	idStr := c.Param("inviteId")
	inviteID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")

	if err := h.inviteService.RejectInvite(uint(inviteID), userID.(uint)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Invite rejected successfully")
}

func (h *GroupInviteHandler) RevokeInvite(c *gin.Context) {
	idStr := c.Param("inviteId")
	inviteID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")

	if err := h.inviteService.RevokeInvite(uint(inviteID), userID.(uint)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Invite revoked successfully")
}
