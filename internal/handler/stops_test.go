package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"tj-routes/internal/models"
	"tj-routes/internal/repository"

	"github.com/stretchr/testify/assert"
)

func TestListStops(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test stops
	stopRepo := repository.NewStopRepository(db)
	stop1 := &models.Stop{
		Name:      "Stop 1",
		Type:      models.StopTypeStop,
		Latitude:  40.7128,
		Longitude: -74.0060,
		City:      "New York",
		Status:    models.StatusActive,
	}
	stopRepo.Create(stop1)

	stop2 := &models.Stop{
		Name:      "Stop 2",
		Type:      models.StopTypeTerminal,
		Latitude:  40.7580,
		Longitude: -73.9855,
		City:      "New York",
		Status:    models.StatusActive,
	}
	stopRepo.Create(stop2)

	t.Run("list stops without auth", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/stops", nil, "")

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		stops := data["stops"].([]interface{})
		assert.GreaterOrEqual(t, len(stops), 2)
	})

	t.Run("list stops with pagination", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/stops?page=1&limit=5", nil, "")

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("list stops with status filter", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/stops?status=active", nil, "")

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("list stops with type filter", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/stops?type=terminal", nil, "")

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestGetStop(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test stop
	stopRepo := repository.NewStopRepository(db)
	stop := &models.Stop{
		Name:      "Test Stop",
		Type:      models.StopTypeStop,
		Latitude:  40.7128,
		Longitude: -74.0060,
		Status:    models.StatusActive,
	}
	stopRepo.Create(stop)

	t.Run("get stop without auth", func(t *testing.T) {
		w := makeRequest(router, "GET", fmt.Sprintf("/api/v1/stops/%d", stop.ID), nil, "")

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("get non-existent stop", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/stops/99999", nil, "")

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestCreateStop(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create admin user
	adminUser := createTestUser(t, db, "admin5@example.com", "admin5", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	// Create regular user
	regularUser := createTestUser(t, db, "user4@example.com", "user4", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(regularUser, cfg)
	assert.NoError(t, err)

	t.Run("create stop as admin", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":      "New Stop",
			"type":      "stop",
			"latitude":  40.7128,
			"longitude": -74.0060,
			"address":   "123 Main St",
			"city":      "New York",
			"status":    "active",
		}

		w := makeRequest(router, "POST", "/api/v1/stops", reqBody, adminToken)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("create stop as regular user (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":      "Unauthorized Stop",
			"type":      "stop",
			"latitude":  40.7128,
			"longitude": -74.0060,
		}

		w := makeRequest(router, "POST", "/api/v1/stops", reqBody, userToken)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("create stop without auth (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":      "Public Stop",
			"type":      "stop",
			"latitude":  40.7128,
			"longitude": -74.0060,
		}

		w := makeRequest(router, "POST", "/api/v1/stops", reqBody, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("create stop with invalid data", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "", // Missing required field
		}

		w := makeRequest(router, "POST", "/api/v1/stops", reqBody, adminToken)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateStop(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test stop
	stopRepo := repository.NewStopRepository(db)
	stop := &models.Stop{
		Name:      "Stop to Update",
		Type:      models.StopTypeStop,
		Latitude:  40.7128,
		Longitude: -74.0060,
		Status:    models.StatusActive,
	}
	stopRepo.Create(stop)

	// Create admin user
	adminUser := createTestUser(t, db, "admin6@example.com", "admin6", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	// Create regular user
	regularUser := createTestUser(t, db, "user5@example.com", "user5", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(regularUser, cfg)
	assert.NoError(t, err)

	t.Run("update stop as admin", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":     "Updated Stop Name",
			"address":  "456 Updated St",
			"latitude": 40.7580,
		}

		w := makeRequest(router, "PUT", fmt.Sprintf("/api/v1/stops/%d", stop.ID), reqBody, adminToken)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("update stop as regular user (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "Unauthorized Update",
		}

		w := makeRequest(router, "PUT", fmt.Sprintf("/api/v1/stops/%d", stop.ID), reqBody, userToken)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestDeleteStop(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create admin user
	adminUser := createTestUser(t, db, "admin7@example.com", "admin7", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	// Create regular user
	regularUser := createTestUser(t, db, "user6@example.com", "user6", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(regularUser, cfg)
	assert.NoError(t, err)

	t.Run("delete stop as admin", func(t *testing.T) {
		// Create stop to delete
		stopRepo := repository.NewStopRepository(db)
		stop := &models.Stop{
			Name:      "Stop to Delete",
			Type:      models.StopTypeStop,
			Latitude:  40.7128,
			Longitude: -74.0060,
			Status:    models.StatusActive,
		}
		stopRepo.Create(stop)

		w := makeRequest(router, "DELETE", fmt.Sprintf("/api/v1/stops/%d", stop.ID), nil, adminToken)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("delete stop as regular user (should fail)", func(t *testing.T) {
		// Create stop to delete
		stopRepo := repository.NewStopRepository(db)
		stop := &models.Stop{
			Name:      "Stop to Delete 2",
			Type:      models.StopTypeStop,
			Latitude:  40.7128,
			Longitude: -74.0060,
			Status:    models.StatusActive,
		}
		stopRepo.Create(stop)

		w := makeRequest(router, "DELETE", fmt.Sprintf("/api/v1/stops/%d", stop.ID), nil, userToken)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

