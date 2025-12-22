package handler

import (
	"net/http"
	"strconv"

	"tj-routes/internal/models"
	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type VehicleHandler struct {
	vehicleService service.VehicleService
}

func NewVehicleHandler(vehicleService service.VehicleService) *VehicleHandler {
	return &VehicleHandler{
		vehicleService: vehicleService,
	}
}

func (h *VehicleHandler) ListVehicles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = models.Status(status)
	}
	if routeID := c.Query("route_id"); routeID != "" {
		id, err := strconv.ParseUint(routeID, 10, 32)
		if err == nil {
			filters["route_id"] = uint(id)
		}
	}

	vehicles, total, err := h.vehicleService.ListVehicles(offset, limit, filters)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"vehicles": vehicles,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

func (h *VehicleHandler) GetVehicle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	vehicle, err := h.vehicleService.GetVehicleByID(uint(id))
	if err != nil {
		NotFound(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, vehicle)
}

func (h *VehicleHandler) CreateVehicle(c *gin.Context) {
	var vehicle models.Vehicle
	if err := c.ShouldBindJSON(&vehicle); err != nil {
		BadRequest(c, err)
		return
	}

	if vehicle.Status == "" {
		vehicle.Status = models.StatusActive
	}

	if err := h.vehicleService.CreateVehicle(&vehicle); err != nil {
		BadRequest(c, err)
		return
	}

	SuccessResponse(c, http.StatusCreated, vehicle)
}

func (h *VehicleHandler) UpdateVehicle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	var vehicle models.Vehicle
	if err := c.ShouldBindJSON(&vehicle); err != nil {
		BadRequest(c, err)
		return
	}

	vehicle.ID = uint(id)
	if err := h.vehicleService.UpdateVehicle(&vehicle); err != nil {
		BadRequest(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, vehicle)
}

func (h *VehicleHandler) DeleteVehicle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	if err := h.vehicleService.DeleteVehicle(uint(id)); err != nil {
		InternalServerError(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Vehicle deleted successfully")
}

