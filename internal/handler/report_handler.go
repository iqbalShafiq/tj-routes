package handler

import (
	"errors"
	"net/http"
	"strconv"

	"tj-routes/internal/models"
	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportService service.ReportService
}

type UpdateReportStatusRequest struct {
	Status     string  `json:"status" binding:"required"`
	AdminNotes *string `json:"admin_notes"`
}

func NewReportHandler(reportService service.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
}

func (h *ReportHandler) ListReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = models.ReportStatus(status)
	}
	if reportType := c.Query("type"); reportType != "" {
		filters["type"] = models.ReportType(reportType)
	}

	// Non-admin users only see their own reports
	userRole, exists := c.Get("user_role")
	if exists && userRole.(models.UserRole) != models.RoleAdmin {
		userID, _ := c.Get("user_id")
		filters["user_id"] = userID.(uint)
	}

	reports, total, err := h.reportService.ListReports(offset, limit, filters)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"reports": reports,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

func (h *ReportHandler) GetReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	report, err := h.reportService.GetReportByID(uint(id))
	if err != nil {
		NotFound(c, err)
		return
	}

	// Non-admin users can only view their own reports
	userRole, exists := c.Get("user_role")
	if exists && userRole.(models.UserRole) != models.RoleAdmin {
		userID, _ := c.Get("user_id")
		if report.UserID != userID.(uint) {
			Forbidden(c, errors.New("access denied: you can only view your own reports"))
			return
		}
	}

	SuccessResponse(c, http.StatusOK, report)
}

func (h *ReportHandler) CreateReport(c *gin.Context) {
	var report models.Report
	if err := c.ShouldBindJSON(&report); err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")
	report.UserID = userID.(uint)

	if err := h.reportService.CreateReport(&report); err != nil {
		BadRequest(c, err)
		return
	}

	SuccessResponse(c, http.StatusCreated, report)
}

func (h *ReportHandler) UpdateReportStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	var req UpdateReportStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	if err := h.reportService.UpdateReportStatus(uint(id), models.ReportStatus(req.Status), req.AdminNotes); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Report status updated successfully")
}

func (h *ReportHandler) DeleteReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	if err := h.reportService.DeleteReport(uint(id)); err != nil {
		InternalServerError(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Report deleted successfully")
}

