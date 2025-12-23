package handler

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"

	"tj-routes/internal/models"
	"tj-routes/internal/repository"
	"tj-routes/internal/service"
	"tj-routes/internal/utils"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportService service.ReportService
	userService   service.UserService
	fileStorage   utils.FileStorage
	reportRepo    repository.ReportRepository
}

type UpdateReportStatusRequest struct {
	Status     string  `json:"status" binding:"required"`
	AdminNotes *string `json:"admin_notes"`
}

func NewReportHandler(reportService service.ReportService, userService service.UserService, fileStorage utils.FileStorage, reportRepo repository.ReportRepository) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		userService:   userService,
		fileStorage:   fileStorage,
		reportRepo:    reportRepo,
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
	if search := c.Query("search"); search != "" {
		filters["search"] = search
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
	var photoFiles []*multipart.FileHeader
	var pdfFiles []*multipart.FileHeader

	// Check if request is multipart/form-data
	contentType := c.GetHeader("Content-Type")
	if contentType == "multipart/form-data" || len(contentType) > 19 && contentType[:19] == "multipart/form-data" {
		// Handle multipart/form-data
		// Extract JSON data from form field "data"
		dataStr := c.PostForm("data")
		if dataStr != "" {
			if err := json.Unmarshal([]byte(dataStr), &report); err != nil {
				BadRequest(c, err)
				return
			}
		} else {
			// If no "data" field, try to bind directly (fallback)
			if err := c.ShouldBind(&report); err != nil {
				BadRequest(c, err)
				return
			}
		}

		// Get multiple photo files
		form, err := c.MultipartForm()
		if err == nil && form != nil {
			// Check for photos (standard multipart) or photos[] (convention)
			if files := form.File["photos"]; len(files) > 0 {
				photoFiles = files
			} else if files := form.File["photos[]"]; len(files) > 0 {
				photoFiles = files
			} else if file := form.File["photo"]; len(file) > 0 {
				// Support single photo as well
				photoFiles = file
			}

			// Get multiple PDF files
			// Check for pdfs (standard multipart) or pdfs[] (convention)
			if files := form.File["pdfs"]; len(files) > 0 {
				pdfFiles = files
			} else if files := form.File["pdfs[]"]; len(files) > 0 {
				pdfFiles = files
			} else if file := form.File["pdf"]; len(file) > 0 {
				// Support single PDF as well
				pdfFiles = file
			}
		}
	} else {
		// Handle JSON-only request
		if err := c.ShouldBindJSON(&report); err != nil {
			BadRequest(c, err)
			return
		}
	}

	// Check if user is authenticated
	userID, exists := c.Get("user_id")
	if exists {
		// Authenticated user - use their ID
		report.UserID = userID.(uint)
	} else {
		// Guest user - use system user ID
		systemUser, err := h.userService.GetSystemUser()
		if err != nil {
			InternalServerError(c, err)
			return
		}
		report.UserID = systemUser.ID
	}

	// Create report first to get ID (we'll update with file URLs after)
	if err := h.reportService.CreateReport(&report); err != nil {
		BadRequest(c, err)
		return
	}

	// Upload photos if provided
	var photoURLs []string
	if len(photoFiles) > 0 {
		for _, file := range photoFiles {
			fileURL, err := h.fileStorage.SaveFile(file, "reports", strconv.FormatUint(uint64(report.ID), 10))
			if err != nil {
				BadRequest(c, err)
				return
			}
			photoURLs = append(photoURLs, fileURL)
		}
	}

	// Upload PDFs if provided
	var pdfURLs []string
	if len(pdfFiles) > 0 {
		for _, file := range pdfFiles {
			fileURL, err := h.fileStorage.SaveFile(file, "reports", strconv.FormatUint(uint64(report.ID), 10))
			if err != nil {
				BadRequest(c, err)
				return
			}
			pdfURLs = append(pdfURLs, fileURL)
		}
	}

	// Update report with file URLs if any files were uploaded
	if len(photoURLs) > 0 || len(pdfURLs) > 0 {
		photoURLsJSON, _ := json.Marshal(photoURLs)
		pdfURLsJSON, _ := json.Marshal(pdfURLs)
		
		photoURLsStr := string(photoURLsJSON)
		pdfURLsStr := string(pdfURLsJSON)
		
		report.PhotoURLs = &photoURLsStr
		report.PDFURLs = &pdfURLsStr
		
		// Update report with file URLs via repository
		if err := h.reportRepo.Update(&report); err != nil {
			InternalServerError(c, err)
			return
		}
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

	// Get report to check for files before deletion
	report, err := h.reportService.GetReportByID(uint(id))
	if err == nil && report != nil {
		// Delete associated photo files
		if report.PhotoURLs != nil {
			var photoURLs []string
			if err := json.Unmarshal([]byte(*report.PhotoURLs), &photoURLs); err == nil {
				for _, url := range photoURLs {
					h.fileStorage.DeleteFile(url)
				}
			}
		}
		// Delete associated PDF files
		if report.PDFURLs != nil {
			var pdfURLs []string
			if err := json.Unmarshal([]byte(*report.PDFURLs), &pdfURLs); err == nil {
				for _, url := range pdfURLs {
					h.fileStorage.DeleteFile(url)
				}
			}
		}
	}

	if err := h.reportService.DeleteReport(uint(id)); err != nil {
		InternalServerError(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Report deleted successfully")
}

