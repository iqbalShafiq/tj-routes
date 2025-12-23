package handler

import (
	"net/http"
	"strconv"

	"tj-routes/internal/models"
	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type BulkUploadHandler struct {
	bulkUploadService service.BulkUploadService
}

func NewBulkUploadHandler(bulkUploadService service.BulkUploadService) *BulkUploadHandler {
	return &BulkUploadHandler{
		bulkUploadService: bulkUploadService,
	}
}

// UploadCSV handles CSV file upload for bulk import
// POST /api/v1/bulk-upload/:entityType
func (h *BulkUploadHandler) UploadCSV(c *gin.Context) {
	if h.bulkUploadService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Bulk upload service is not available. Please check Redis connection and restart the application.",
		})
		return
	}

	entityType := c.Param("entityType")

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		BadRequest(c, err)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Upload CSV
	uploadLog, err := h.bulkUploadService.UploadCSV(entityType, file, userID.(uint))
	if err != nil {
		BadRequest(c, err)
		return
	}

	SuccessResponse(c, http.StatusAccepted, uploadLog)
}

// GetUploadStatus gets the status of a bulk upload
// GET /api/v1/bulk-upload/:id
func (h *BulkUploadHandler) GetUploadStatus(c *gin.Context) {
	if h.bulkUploadService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Bulk upload service is not available. Please check Redis connection and restart the application.",
		})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	uploadLog, err := h.bulkUploadService.GetUploadStatus(uint(id))
	if err != nil {
		NotFound(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, uploadLog)
}

// ListUploads lists all bulk uploads with optional filters
// GET /api/v1/bulk-upload
func (h *BulkUploadHandler) ListUploads(c *gin.Context) {
	if h.bulkUploadService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Bulk upload service is not available. Please check Redis connection and restart the application.",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	filters := make(map[string]interface{})
	if entityType := c.Query("entity_type"); entityType != "" {
		filters["entity_type"] = models.BulkUploadEntityType(entityType)
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = models.BulkUploadStatus(status)
	}
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			filters["user_id"] = uint(userID)
		}
	}

	uploads, total, err := h.bulkUploadService.ListUploads(offset, limit, filters)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"uploads": uploads,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

