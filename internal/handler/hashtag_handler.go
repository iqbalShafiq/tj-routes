package handler

import (
	"net/http"
	"strconv"

	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type HashtagHandler struct {
	hashtagService service.HashtagService
}

func NewHashtagHandler(hashtagService service.HashtagService) *HashtagHandler {
	return &HashtagHandler{
		hashtagService: hashtagService,
	}
}

func (h *HashtagHandler) GetTrending(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	hashtags, err := h.hashtagService.GetTrending(limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"hashtags": hashtags,
	})
}

func (h *HashtagHandler) GetReportsByHashtag(c *gin.Context) {
	hashtagName := c.Param("name")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	reports, total, err := h.hashtagService.GetReportsByHashtag(hashtagName, offset, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"reports": reports,
		"total":   total,
		"page":    page,
		"limit":   limit,
		"hashtag": hashtagName,
	})
}

func (h *HashtagHandler) SearchHashtags(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		BadRequest(c, nil)
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	hashtags, err := h.hashtagService.SearchHashtags(query, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"hashtags": hashtags,
		"query":    query,
	})
}

