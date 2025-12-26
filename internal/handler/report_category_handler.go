package handler

import (
	"net/http"

	"tj-routes/internal/models"
	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type ReportCategoryHandler struct {
	categoryService service.ReportCategoryService
}

func NewReportCategoryHandler(categoryService service.ReportCategoryService) *ReportCategoryHandler {
	return &ReportCategoryHandler{
		categoryService: categoryService,
	}
}

type CreateCategoryRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

func (h *ReportCategoryHandler) ListCategories(c *gin.Context) {
	categories, err := h.categoryService.ListCategories()
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"categories": categories,
	})
}

func (h *ReportCategoryHandler) CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	category := &models.ReportCategory{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.categoryService.CreateCategory(category); err != nil {
		BadRequest(c, err)
		return
	}

	SuccessResponse(c, http.StatusCreated, category)
}

