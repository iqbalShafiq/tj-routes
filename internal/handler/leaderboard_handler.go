package handler

import (
	"net/http"
	"strconv"

	"tj-routes/internal/models"
	"tj-routes/internal/repository"
	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type LeaderboardHandler struct {
	userRepo        repository.UserRepository
	badgeService    service.BadgeService
	reputationService service.ReputationService
}

func NewLeaderboardHandler(
	userRepo repository.UserRepository,
	badgeService service.BadgeService,
	reputationService service.ReputationService,
) *LeaderboardHandler {
	return &LeaderboardHandler{
		userRepo:          userRepo,
		badgeService:      badgeService,
		reputationService: reputationService,
	}
}

func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	users, err := h.userRepo.Leaderboard(limit)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"leaderboard": users,
		"limit":       limit,
	})
}

func (h *LeaderboardHandler) GetUserProfile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	user, err := h.userRepo.FindByID(uint(id))
	if err != nil {
		NotFound(c, err)
		return
	}

	badges, err := h.badgeService.GetUserBadges(uint(id))
	if err != nil {
		// Continue even if badges fail to load
		badges = []models.UserBadge{}
	}

	points, level, _ := h.reputationService.GetUserReputation(uint(id))

	SuccessResponse(c, http.StatusOK, gin.H{
		"user":              user,
		"reputation_points": points,
		"level":             level,
		"badges":            badges,
	})
}

func (h *LeaderboardHandler) GetAllBadges(c *gin.Context) {
	badges, err := h.badgeService.GetAllBadges()
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"badges": badges,
	})
}

