package handler

import (
	"net/http"
	"strconv"

	"tj-routes/internal/models"
	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type CheckInHandler struct {
	checkInService service.CheckInService
}

func NewCheckInHandler(checkInService service.CheckInService) *CheckInHandler {
	return &CheckInHandler{
		checkInService: checkInService,
	}
}

func (h *CheckInHandler) CreateCheckIn(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, nil)
		return
	}
	userID := userIDVal.(uint)

	var req service.CreateCheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	checkIn, err := h.checkInService.CreateCheckIn(c.Request.Context(), userID, &req)
	if err != nil {
		switch err {
		case service.ErrActiveCheckInExists:
			ErrorResponse(c, http.StatusConflict, err)
		case service.ErrRouteNotFound:
			NotFound(c, err)
		case service.ErrStopNotFound:
			NotFound(c, err)
		default:
			InternalServerError(c, err)
		}
		return
	}

	SuccessResponse(c, http.StatusCreated, checkIn)
}

func (h *CheckInHandler) CompleteCheckIn(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, nil)
		return
	}
	userID := userIDVal.(uint)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	var req service.CompleteCheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	checkIn, err := h.checkInService.CompleteCheckIn(c.Request.Context(), userID, uint(id), &req)
	if err != nil {
		handleCheckInError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, checkIn)
}

func (h *CheckInHandler) UpdateCheckIn(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, nil)
		return
	}
	userID := userIDVal.(uint)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	var req service.UpdateCheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	checkIn, err := h.checkInService.UpdateCheckIn(c.Request.Context(), userID, uint(id), &req)
	if err != nil {
		handleCheckInError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, checkIn)
}

func (h *CheckInHandler) GetCheckIn(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, nil)
		return
	}
	userID := userIDVal.(uint)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	checkIn, err := h.checkInService.GetCheckIn(c.Request.Context(), userID, uint(id))
	if err != nil {
		handleCheckInError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, checkIn)
}

func (h *CheckInHandler) ListCheckIns(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, nil)
		return
	}
	userID := userIDVal.(uint)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	var status *models.CheckInStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s := models.CheckInStatus(statusStr)
		status = &s
	}

	checkIns, total, err := h.checkInService.ListCheckIns(c.Request.Context(), userID, page, limit, status)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"check_ins": checkIns,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

func (h *CheckInHandler) DeleteCheckIn(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, nil)
		return
	}
	userID := userIDVal.(uint)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	if err := h.checkInService.DeleteCheckIn(c.Request.Context(), userID, uint(id)); err != nil {
		handleCheckInError(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Check-in deleted successfully")
}

func (h *CheckInHandler) GetStats(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, nil)
		return
	}
	userID := userIDVal.(uint)

	stats, err := h.checkInService.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, stats)
}

func handleCheckInError(c *gin.Context, err error) {
	switch err {
	case service.ErrNotAuthorized:
		Forbidden(c, err)
	case service.ErrActiveCheckInExists:
		ErrorResponse(c, http.StatusConflict, err)
	case service.ErrCheckInNotInProgress:
		ErrorResponse(c, http.StatusBadRequest, err)
	case service.ErrRouteNotFound, service.ErrStopNotFound:
		NotFound(c, err)
	default:
		NotFound(c, err)
	}
}
