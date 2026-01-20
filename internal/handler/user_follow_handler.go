package handler

import (
	"net/http"
	"strconv"

	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type UserFollowHandler struct {
	userFollowService service.UserFollowService
}

func NewUserFollowHandler(userFollowService service.UserFollowService) *UserFollowHandler {
	return &UserFollowHandler{
		userFollowService: userFollowService,
	}
}

func (h *UserFollowHandler) FollowUser(c *gin.Context) {
	userID, _ := c.Get("user_id")
	followerID := userID.(uint)

	followingIDStr := c.Param("id")
	followingID, err := strconv.ParseUint(followingIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	if err := h.userFollowService.Follow(followerID, uint(followingID)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Successfully followed user")
}

func (h *UserFollowHandler) UnfollowUser(c *gin.Context) {
	userID, _ := c.Get("user_id")
	followerID := userID.(uint)

	followingIDStr := c.Param("id")
	followingID, err := strconv.ParseUint(followingIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	if err := h.userFollowService.Unfollow(followerID, uint(followingID)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Successfully unfollowed user")
}

func (h *UserFollowHandler) GetFollowers(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	followers, total, err := h.userFollowService.GetFollowers(uint(userID), offset, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"followers": followers,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

func (h *UserFollowHandler) GetFollowing(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	following, total, err := h.userFollowService.GetFollowing(uint(userID), offset, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"following": following,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

func (h *UserFollowHandler) GetFollowStatus(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, nil)
		return
	}
	followerID := currentUserID.(uint)

	userIDStr := c.Param("id")
	followingID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	isFollowing, err := h.userFollowService.IsFollowing(followerID, uint(followingID))
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"is_following": isFollowing,
	})
}

func (h *UserFollowHandler) GetFriends(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	friends, total, err := h.userFollowService.GetFriends(uint(userID), offset, limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"friends": friends,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

