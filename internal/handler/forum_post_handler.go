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

type ForumPostHandler struct {
	forumPostService service.ForumPostService
	fileStorage      utils.FileStorage
}

func NewForumPostHandler(forumPostService service.ForumPostService, fileStorage utils.FileStorage) *ForumPostHandler {
	return &ForumPostHandler{
		forumPostService: forumPostService,
		fileStorage:      fileStorage,
	}
}

type CreateForumPostRequest struct {
	PostType       string  `json:"post_type" binding:"required"`
	Title          string  `json:"title" binding:"required"`
	Content        string  `json:"content" binding:"required"`
	LinkedReportID *uint   `json:"linked_report_id,omitempty"`
}

type UpdateForumPostRequest struct {
	Title    string  `json:"title" binding:"required"`
	Content  string  `json:"content" binding:"required"`
	PostType *string `json:"post_type,omitempty"`
}

func (h *ForumPostHandler) ListPosts(c *gin.Context) {
	forumIDStr := c.Param("id")
	forumID, err := strconv.ParseUint(forumIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	filters := make(map[string]interface{})
	if postType := c.Query("post_type"); postType != "" {
		filters["post_type"] = models.PostType(postType)
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	posts, total, err := h.forumPostService.ListPosts(uint(forumID), offset, limit, filters)
	if err != nil {
		NotFound(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"posts": posts,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *ForumPostHandler) GetPost(c *gin.Context) {
	forumIDStr := c.Param("id")
	postIDStr := c.Param("postId")
	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	// Verify forum ID matches (optional validation)
	_, err = strconv.ParseUint(forumIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	post, err := h.forumPostService.GetPostByID(uint(postID))
	if err != nil {
		NotFound(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, post)
}

func (h *ForumPostHandler) CreatePost(c *gin.Context) {
	forumIDStr := c.Param("id")
	forumID, err := strconv.ParseUint(forumIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")

	var post models.ForumPost
	var photoFiles []*multipart.FileHeader
	var pdfFiles []*multipart.FileHeader

	// Check if request is multipart/form-data
	contentType := c.GetHeader("Content-Type")
	if contentType == "multipart/form-data" || len(contentType) > 19 && contentType[:19] == "multipart/form-data" {
		// Handle multipart/form-data
		dataStr := c.PostForm("data")
		if dataStr != "" {
			if err := json.Unmarshal([]byte(dataStr), &post); err != nil {
				BadRequest(c, err)
				return
			}
		} else {
			if err := c.ShouldBind(&post); err != nil {
				BadRequest(c, err)
				return
			}
		}

		// Get multiple photo files
		form, err := c.MultipartForm()
		if err == nil && form != nil {
			if files := form.File["photos"]; len(files) > 0 {
				photoFiles = files
			} else if files := form.File["photos[]"]; len(files) > 0 {
				photoFiles = files
			} else if file := form.File["photo"]; len(file) > 0 {
				photoFiles = file
			}

			if files := form.File["pdfs"]; len(files) > 0 {
				pdfFiles = files
			} else if files := form.File["pdfs[]"]; len(files) > 0 {
				pdfFiles = files
			} else if file := form.File["pdf"]; len(file) > 0 {
				pdfFiles = file
			}
		}
	} else {
		// Handle JSON-only request
		var req CreateForumPostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequest(c, err)
			return
		}
		post.PostType = models.PostType(req.PostType)
		post.Title = req.Title
		post.Content = req.Content
		post.LinkedReportID = req.LinkedReportID
	}

	post.ForumID = uint(forumID)

	// Create post first
	if err := h.forumPostService.CreatePost(&post, userID.(uint)); err != nil {
		BadRequest(c, err)
		return
	}

	// Upload photos if provided
	var photoURLs []string
	if len(photoFiles) > 0 {
		for _, file := range photoFiles {
			fileURL, err := h.fileStorage.SaveFile(file, "forum-posts", strconv.FormatUint(uint64(post.ID), 10))
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
			fileURL, err := h.fileStorage.SaveFile(file, "forum-posts", strconv.FormatUint(uint64(post.ID), 10))
			if err != nil {
				BadRequest(c, err)
				return
			}
			pdfURLs = append(pdfURLs, fileURL)
		}
	}

	// Update post with file URLs if any
	if len(photoURLs) > 0 || len(pdfURLs) > 0 {
		if err := h.forumPostService.UpdatePostWithFiles(post.ID, userID.(uint), photoURLs, pdfURLs); err != nil {
			// Log error but don't fail the request - files are uploaded
		}
	}

	// Reload post with relationships
	reloadedPost, _ := h.forumPostService.GetPostByID(post.ID)
	if reloadedPost != nil {
		SuccessResponse(c, http.StatusCreated, reloadedPost)
		return
	}

	SuccessResponse(c, http.StatusCreated, post)
}

func (h *ForumPostHandler) UpdatePost(c *gin.Context) {
	forumIDStr := c.Param("id")
	postIDStr := c.Param("postId")
	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	_, err = strconv.ParseUint(forumIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")

	var req UpdateForumPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	var postType *models.PostType
	if req.PostType != nil {
		pt := models.PostType(*req.PostType)
		postType = &pt
	}

	if err := h.forumPostService.UpdatePost(uint(postID), userID.(uint), req.Title, req.Content, postType); err != nil {
		BadRequest(c, err)
		return
	}

	post, _ := h.forumPostService.GetPostByID(uint(postID))
	SuccessResponse(c, http.StatusOK, post)
}

func (h *ForumPostHandler) DeletePost(c *gin.Context) {
	forumIDStr := c.Param("id")
	postIDStr := c.Param("postId")
	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	_, err = strconv.ParseUint(forumIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")
	isAdmin := userRole.(models.UserRole) == models.RoleAdmin

	if err := h.forumPostService.DeletePost(uint(postID), userID.(uint), isAdmin); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Forum post deleted successfully")
}

func (h *ForumPostHandler) PinPost(c *gin.Context) {
	postIDStr := c.Param("postId")
	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	if err := h.forumPostService.PinPost(uint(postID)); err != nil {
		BadRequest(c, err)
		return
	}

	post, _ := h.forumPostService.GetPostByID(uint(postID))
	SuccessResponse(c, http.StatusOK, post)
}

func (h *ForumPostHandler) UnpinPost(c *gin.Context) {
	postIDStr := c.Param("postId")
	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	if err := h.forumPostService.UnpinPost(uint(postID)); err != nil {
		BadRequest(c, err)
		return
	}

	post, _ := h.forumPostService.GetPostByID(uint(postID))
	SuccessResponse(c, http.StatusOK, post)
}

