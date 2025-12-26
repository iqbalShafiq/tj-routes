package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"tj-routes/internal/models"
	"tj-routes/internal/repository"

	"github.com/stretchr/testify/assert"
)

func TestListReportCategories(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Seed some categories
	categoryRepo := repository.NewReportCategoryRepository(db)
	category1 := &models.ReportCategory{
		Name: "Crash",
	}
	categoryRepo.Create(category1)

	category2 := &models.ReportCategory{
		Name: "Trash",
	}
	categoryRepo.Create(category2)

	t.Run("list categories as guest", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/report-categories", nil, "")

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		categories := data["categories"].([]interface{})
		assert.GreaterOrEqual(t, len(categories), 2)
	})
}

func TestCreateReportCategory(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create admin user
	admin := createTestUser(t, db, "admin@example.com", "admin", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(admin, cfg)
	assert.NoError(t, err)

	t.Run("create category as admin", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":        "Test Category",
			"description": "Test description",
		}

		w := makeRequest(router, "POST", "/api/v1/report-categories", reqBody, adminToken)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("create category without auth", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "Unauthorized Category",
		}

		w := makeRequest(router, "POST", "/api/v1/report-categories", reqBody, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

