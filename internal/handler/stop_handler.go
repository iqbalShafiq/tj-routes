package handler

import (
	"net/http"
	"strconv"

	"tj-routes/internal/models"
	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type StopHandler struct {
	stopService service.StopService
}

func NewStopHandler(stopService service.StopService) *StopHandler {
	return &StopHandler{
		stopService: stopService,
	}
}

func (h *StopHandler) ListStops(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = models.Status(status)
	}
	if stopType := c.Query("type"); stopType != "" {
		filters["type"] = models.StopType(stopType)
	}
	if city := c.Query("city"); city != "" {
		filters["city"] = city
	}

	stops, total, err := h.stopService.ListStops(offset, limit, filters)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"stops": stops,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *StopHandler) GetStop(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	stop, err := h.stopService.GetStopByID(uint(id))
	if err != nil {
		NotFound(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, stop)
}

func (h *StopHandler) CreateStop(c *gin.Context) {
	var stop models.Stop
	if err := c.ShouldBindJSON(&stop); err != nil {
		BadRequest(c, err)
		return
	}

	if err := h.stopService.CreateStop(&stop); err != nil {
		BadRequest(c, err)
		return
	}

	SuccessResponse(c, http.StatusCreated, stop)
}

func (h *StopHandler) UpdateStop(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	var stop models.Stop
	if err := c.ShouldBindJSON(&stop); err != nil {
		BadRequest(c, err)
		return
	}

	stop.ID = uint(id)
	if err := h.stopService.UpdateStop(&stop); err != nil {
		BadRequest(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, stop)
}

func (h *StopHandler) DeleteStop(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	if err := h.stopService.DeleteStop(uint(id)); err != nil {
		InternalServerError(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Stop deleted successfully")
}

