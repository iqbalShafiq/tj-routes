package handler

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"

	"tj-routes/internal/models"
	"tj-routes/internal/service"
	"tj-routes/internal/utils"

	"github.com/gin-gonic/gin"
)

type VehicleHandler struct {
	vehicleService service.VehicleService
	fileStorage    utils.FileStorage
}

func NewVehicleHandler(vehicleService service.VehicleService, fileStorage utils.FileStorage) *VehicleHandler {
	return &VehicleHandler{
		vehicleService: vehicleService,
		fileStorage:    fileStorage,
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
	if search := c.Query("search"); search != "" {
		filters["search"] = search
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
	var file *multipart.FileHeader

	// Check if request is multipart/form-data
	contentType := c.GetHeader("Content-Type")
	if contentType == "multipart/form-data" || len(contentType) > 19 && contentType[:19] == "multipart/form-data" {
		// Handle multipart/form-data
		// Extract JSON data from form field "data"
		dataStr := c.PostForm("data")
		if dataStr != "" {
			if err := json.Unmarshal([]byte(dataStr), &vehicle); err != nil {
				BadRequest(c, err)
				return
			}
		} else {
			// If no "data" field, try to bind directly (fallback)
			if err := c.ShouldBind(&vehicle); err != nil {
				BadRequest(c, err)
				return
			}
		}

		// Get photo file (don't save yet, need ID first)
		file, _ = c.FormFile("photo")
	} else {
		// Handle JSON-only request
		if err := c.ShouldBindJSON(&vehicle); err != nil {
			BadRequest(c, err)
			return
		}
	}

	if vehicle.Status == "" {
		vehicle.Status = models.StatusActive
	}

	// Create vehicle first to get ID
	if err := h.vehicleService.CreateVehicle(&vehicle); err != nil {
		BadRequest(c, err)
		return
	}

	// If file was uploaded, save it now with the correct ID
	if file != nil {
		fileURL, err := h.fileStorage.SaveFile(file, "vehicles", strconv.FormatUint(uint64(vehicle.ID), 10))
		if err != nil {
			BadRequest(c, err)
			return
		}
		vehicle.PhotoURL = &fileURL
		// Update vehicle with photo URL
		if err := h.vehicleService.UpdateVehicle(&vehicle); err != nil {
			BadRequest(c, err)
			return
		}
	}

	SuccessResponse(c, http.StatusCreated, vehicle)
}

func (h *VehicleHandler) UpdateVehicle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	// Get existing vehicle to check for old photo
	existingVehicle, err := h.vehicleService.GetVehicleByID(uint(id))
	if err != nil {
		NotFound(c, err)
		return
	}

	var vehicle models.Vehicle
	var photoURL *string

	// Check if request is multipart/form-data
	contentType := c.GetHeader("Content-Type")
	if contentType == "multipart/form-data" || len(contentType) > 19 && contentType[:19] == "multipart/form-data" {
		// Handle multipart/form-data
		// Extract JSON data from form field "data"
		dataStr := c.PostForm("data")
		if dataStr != "" {
			if err := json.Unmarshal([]byte(dataStr), &vehicle); err != nil {
				BadRequest(c, err)
				return
			}
		} else {
			// If no "data" field, try to bind directly (fallback)
			if err := c.ShouldBind(&vehicle); err != nil {
				BadRequest(c, err)
				return
			}
		}

		// Handle photo upload
		file, err := c.FormFile("photo")
		if err == nil && file != nil {
			// Delete old photo if exists
			if existingVehicle.PhotoURL != nil {
				h.fileStorage.DeleteFile(*existingVehicle.PhotoURL)
			}

			// Validate and save new file
			fileURL, err := h.fileStorage.SaveFile(file, "vehicles", strconv.FormatUint(uint64(id), 10))
			if err != nil {
				BadRequest(c, err)
				return
			}
			photoURL = &fileURL
		}
	} else {
		// Handle JSON-only request
		if err := c.ShouldBindJSON(&vehicle); err != nil {
			BadRequest(c, err)
			return
		}
	}

	vehicle.ID = uint(id)
	
	// Set photo URL if uploaded, otherwise keep existing
	if photoURL != nil {
		vehicle.PhotoURL = photoURL
	} else if existingVehicle.PhotoURL != nil {
		vehicle.PhotoURL = existingVehicle.PhotoURL
	}

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

	// Get vehicle to check for photo before deletion
	vehicle, err := h.vehicleService.GetVehicleByID(uint(id))
	if err == nil && vehicle != nil && vehicle.PhotoURL != nil {
		// Delete associated photo file
		h.fileStorage.DeleteFile(*vehicle.PhotoURL)
	}

	if err := h.vehicleService.DeleteVehicle(uint(id)); err != nil {
		InternalServerError(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Vehicle deleted successfully")
}

