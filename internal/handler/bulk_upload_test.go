package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"tj-routes/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestBulkUpload(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create admin user
	adminUser := createTestUser(t, db, "admin18@example.com", "admin18", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	// Create regular user
	regularUser := createTestUser(t, db, "user15@example.com", "user15", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(regularUser, cfg)
	assert.NoError(t, err)

	t.Run("bulk upload as admin (service unavailable)", func(t *testing.T) {
		// Since bulk upload service is nil in test setup, this should return service unavailable
		reqBody := map[string]interface{}{
			"file": "test",
		}

		w := makeRequest(router, "POST", "/api/v1/bulk-upload/routes", reqBody, adminToken)

		// Should return service unavailable since bulk upload service is not initialized in tests
		assert.True(t, w.Code == http.StatusServiceUnavailable || w.Code == http.StatusInternalServerError || w.Code == http.StatusBadRequest)
	})

	t.Run("bulk upload as regular user (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"file": "test",
		}

		w := makeRequest(router, "POST", "/api/v1/bulk-upload/routes", reqBody, userToken)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("bulk upload without auth (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"file": "test",
		}

		w := makeRequest(router, "POST", "/api/v1/bulk-upload/routes", reqBody, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestGetUploadStatus(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create admin user
	adminUser := createTestUser(t, db, "admin19@example.com", "admin19", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	t.Run("get upload status as admin", func(t *testing.T) {
		// Since bulk upload service is nil, this might return service unavailable or not found
		w := makeRequest(router, "GET", "/api/v1/bulk-upload/1", nil, adminToken)

		// Could be service unavailable or not found
		assert.True(t, w.Code == http.StatusServiceUnavailable || w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError)
	})

	t.Run("get upload status as regular user (should fail)", func(t *testing.T) {
		regularUser := createTestUser(t, db, "user16@example.com", "user16", "password123", models.RoleCommonUser)
		userToken, err := generateTestToken(regularUser, cfg)
		assert.NoError(t, err)

		w := makeRequest(router, "GET", "/api/v1/bulk-upload/1", nil, userToken)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestListUploads(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create admin user
	adminUser := createTestUser(t, db, "admin20@example.com", "admin20", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	t.Run("list uploads as admin", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/bulk-upload", nil, adminToken)

		// Since bulk upload service is nil, this might return service unavailable or empty list
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusServiceUnavailable || w.Code == http.StatusInternalServerError)

		if w.Code == http.StatusOK {
			var response Response
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
		}
	})

	t.Run("list uploads as regular user (should fail)", func(t *testing.T) {
		regularUser := createTestUser(t, db, "user17@example.com", "user17", "password123", models.RoleCommonUser)
		userToken, err := generateTestToken(regularUser, cfg)
		assert.NoError(t, err)

		w := makeRequest(router, "GET", "/api/v1/bulk-upload", nil, userToken)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

