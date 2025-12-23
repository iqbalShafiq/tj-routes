package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"tj-routes/internal/models"
	"tj-routes/internal/repository"

	"github.com/stretchr/testify/assert"
)

func TestGetLeaderboard(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test users with different reputation points
	user1 := createTestUser(t, db, "leader1@example.com", "leader1", "password123", models.RoleCommonUser)
	user2 := createTestUser(t, db, "leader2@example.com", "leader2", "password123", models.RoleCommonUser)

	// Update reputation points directly (since we don't have a helper)
	userRepo := repository.NewUserRepository(db)
	user1.ReputationPoints = 100
	userRepo.Update(user1)
	user2.ReputationPoints = 50
	userRepo.Update(user2)

	// Create tokens
	user1Token, err := generateTestToken(user1, cfg)
	assert.NoError(t, err)

	t.Run("get leaderboard", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/leaderboard", nil, user1Token)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		assert.NotNil(t, data["leaderboard"])
	})

	t.Run("get leaderboard with limit", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/leaderboard?limit=5", nil, user1Token)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("get leaderboard without auth (should fail)", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/leaderboard", nil, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestGetAllBadges(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test user
	user := createTestUser(t, db, "badgeviewer@example.com", "badgeviewer", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(user, cfg)
	assert.NoError(t, err)

	t.Run("get all badges", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/badges", nil, userToken)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		assert.NotNil(t, data["badges"])
	})

	t.Run("get badges without auth (should fail)", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/badges", nil, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

