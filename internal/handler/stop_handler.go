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

type StopHandler struct {
	stopService service.StopService
	fileStorage utils.FileStorage
}

func NewStopHandler(stopService service.StopService, fileStorage utils.FileStorage) *StopHandler {
	return &StopHandler{
		stopService: stopService,
		fileStorage: fileStorage,
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
	var file *multipart.FileHeader

	// Check if request is multipart/form-data
	contentType := c.GetHeader("Content-Type")
	if contentType == "multipart/form-data" || len(contentType) > 19 && contentType[:19] == "multipart/form-data" {
		// Handle multipart/form-data
		// Extract JSON data from form field "data"
		dataStr := c.PostForm("data")
		if dataStr != "" {
			if err := json.Unmarshal([]byte(dataStr), &stop); err != nil {
				BadRequest(c, err)
				return
			}
		} else {
			// If no "data" field, try to bind directly (fallback)
			if err := c.ShouldBind(&stop); err != nil {
				BadRequest(c, err)
				return
			}
		}

		// Get photo file (don't save yet, need ID first)
		file, _ = c.FormFile("photo")
	} else {
		// Handle JSON-only request
		if err := c.ShouldBindJSON(&stop); err != nil {
			BadRequest(c, err)
			return
		}
	}

	// Create stop first to get ID
	if err := h.stopService.CreateStop(&stop); err != nil {
		BadRequest(c, err)
		return
	}

	// If file was uploaded, save it now with the correct ID
	if file != nil {
		fileURL, err := h.fileStorage.SaveFile(file, "stops", strconv.FormatUint(uint64(stop.ID), 10))
		if err != nil {
			// If file save fails, delete the created stop or just log the error
			// For now, we'll continue but the stop won't have a photo
			BadRequest(c, err)
			return
		}
		stop.PhotoURL = &fileURL
		// Update stop with photo URL
		if err := h.stopService.UpdateStop(&stop); err != nil {
			BadRequest(c, err)
			return
		}
	}

	SuccessResponse(c, http.StatusCreated, stop)
}

func (h *StopHandler) UpdateStop(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	// Get existing stop to check for old photo
	existingStop, err := h.stopService.GetStopByID(uint(id))
	if err != nil {
		NotFound(c, err)
		return
	}

	var stop models.Stop
	var photoURL *string

	// Check if request is multipart/form-data
	contentType := c.GetHeader("Content-Type")
	if contentType == "multipart/form-data" || len(contentType) > 19 && contentType[:19] == "multipart/form-data" {
		// Handle multipart/form-data
		// Extract JSON data from form field "data"
		dataStr := c.PostForm("data")
		if dataStr != "" {
			if err := json.Unmarshal([]byte(dataStr), &stop); err != nil {
				BadRequest(c, err)
				return
			}
		} else {
			// If no "data" field, try to bind directly (fallback)
			if err := c.ShouldBind(&stop); err != nil {
				BadRequest(c, err)
				return
			}
		}

		// Handle photo upload
		file, err := c.FormFile("photo")
		if err == nil && file != nil {
			// Delete old photo if exists
			if existingStop.PhotoURL != nil {
				h.fileStorage.DeleteFile(*existingStop.PhotoURL)
			}

			// Validate and save new file
			fileURL, err := h.fileStorage.SaveFile(file, "stops", strconv.FormatUint(uint64(id), 10))
			if err != nil {
				BadRequest(c, err)
				return
			}
			photoURL = &fileURL
		}
	} else {
		// Handle JSON-only request
		if err := c.ShouldBindJSON(&stop); err != nil {
			BadRequest(c, err)
			return
		}
	}

	stop.ID = uint(id)
	
	// Set photo URL if uploaded, otherwise keep existing
	if photoURL != nil {
		stop.PhotoURL = photoURL
	} else if existingStop.PhotoURL != nil {
		stop.PhotoURL = existingStop.PhotoURL
	}

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

	// Get stop to check for photo before deletion
	stop, err := h.stopService.GetStopByID(uint(id))
	if err == nil && stop != nil && stop.PhotoURL != nil {
		// Delete associated photo file
		h.fileStorage.DeleteFile(*stop.PhotoURL)
	}

	if err := h.stopService.DeleteStop(uint(id)); err != nil {
		InternalServerError(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Stop deleted successfully")
}

