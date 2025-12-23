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

func TestListRoutes(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test data
	stopRepo := repository.NewStopRepository(db)
	routeRepo := repository.NewRouteRepository(db)
	routeStopRepo := repository.NewRouteStopRepository(db)

	stop1 := &models.Stop{
		Name:      "Stop 1",
		Type:      models.StopTypeStop,
		Latitude:  40.7128,
		Longitude: -74.0060,
		Status:    models.StatusActive,
	}
	stopRepo.Create(stop1)

	stop2 := &models.Stop{
		Name:      "Stop 2",
		Type:      models.StopTypeStop,
		Latitude:  40.7580,
		Longitude: -73.9855,
		Status:    models.StatusActive,
	}
	stopRepo.Create(stop2)

	route := &models.Route{
		RouteNumber: "R1",
		Name:        "Route 1",
		Description: "Test Route",
		Status:      models.StatusActive,
	}
	routeRepo.Create(route)

	routeStopRepo.Create(&models.RouteStop{
		RouteID: route.ID,
		StopID:  stop1.ID,
		SequenceOrder: 1,
	})
	routeStopRepo.Create(&models.RouteStop{
		RouteID: route.ID,
		StopID:  stop2.ID,
		SequenceOrder: 2,
	})

	t.Run("list routes without auth", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/routes", nil, "")

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		routes := data["routes"].([]interface{})
		assert.GreaterOrEqual(t, len(routes), 1)
	})

	t.Run("list routes with pagination", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/routes?page=1&limit=5", nil, "")

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("list routes with status filter", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/routes?status=active", nil, "")

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestGetRoute(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test data
	stopRepo := repository.NewStopRepository(db)
	routeRepo := repository.NewRouteRepository(db)

	stop1 := &models.Stop{
		Name:      "Stop 1",
		Type:      models.StopTypeStop,
		Latitude:  40.7128,
		Longitude: -74.0060,
		Status:    models.StatusActive,
	}
	stopRepo.Create(stop1)

	route := &models.Route{
		RouteNumber: "R2",
		Name:        "Route 2",
		Description: "Test Route 2",
		Status:      models.StatusActive,
	}
	routeRepo.Create(route)

	t.Run("get route without auth", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/routes/"+toString(route.ID), nil, "")

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("get non-existent route", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/routes/99999", nil, "")

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestCreateRoute(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test stops
	stopRepo := repository.NewStopRepository(db)
	stop1 := &models.Stop{
		Name:      "Stop A",
		Type:      models.StopTypeStop,
		Latitude:  40.7128,
		Longitude: -74.0060,
		Status:    models.StatusActive,
	}
	stopRepo.Create(stop1)

	stop2 := &models.Stop{
		Name:      "Stop B",
		Type:      models.StopTypeStop,
		Latitude:  40.7580,
		Longitude: -73.9855,
		Status:    models.StatusActive,
	}
	stopRepo.Create(stop2)

	// Create admin user
	adminUser := createTestUser(t, db, "admin@example.com", "admin", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	// Create regular user
	regularUser := createTestUser(t, db, "user@example.com", "user", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(regularUser, cfg)
	assert.NoError(t, err)

	t.Run("create route as admin", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"route_number": "R10",
			"name":         "Route 10",
			"description":  "Test Route 10",
			"status":       "active",
			"stop_ids":     []uint{stop1.ID, stop2.ID},
		}

		w := makeRequest(router, "POST", "/api/v1/routes", reqBody, adminToken)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("create route as regular user (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"route_number": "R11",
			"name":         "Route 11",
			"stop_ids":     []uint{stop1.ID, stop2.ID},
		}

		w := makeRequest(router, "POST", "/api/v1/routes", reqBody, userToken)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("create route without auth (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"route_number": "R12",
			"name":         "Route 12",
			"stop_ids":     []uint{stop1.ID, stop2.ID},
		}

		w := makeRequest(router, "POST", "/api/v1/routes", reqBody, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("create route with invalid stop_ids", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"route_number": "R13",
			"name":         "Route 13",
			"stop_ids":     []uint{99999}, // Non-existent stop
		}

		w := makeRequest(router, "POST", "/api/v1/routes", reqBody, adminToken)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create route with insufficient stops", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"route_number": "R14",
			"name":         "Route 14",
			"stop_ids":     []uint{stop1.ID}, // Only one stop, need at least 2
		}

		w := makeRequest(router, "POST", "/api/v1/routes", reqBody, adminToken)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateRoute(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test route
	routeRepo := repository.NewRouteRepository(db)
	route := &models.Route{
		RouteNumber: "R3",
		Name:        "Route 3",
		Status:      models.StatusActive,
	}
	routeRepo.Create(route)

	// Create admin user
	adminUser := createTestUser(t, db, "admin2@example.com", "admin2", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	// Create regular user
	regularUser := createTestUser(t, db, "user2@example.com", "user2", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(regularUser, cfg)
	assert.NoError(t, err)

	t.Run("update route as admin", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"route_number": "R3-UPDATED",
			"name":         "Route 3 Updated",
			"description":  "Updated description",
		}

		w := makeRequest(router, "PUT", "/api/v1/routes/"+toString(route.ID), reqBody, adminToken)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("update route as regular user (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "Route 3 Updated Again",
		}

		w := makeRequest(router, "PUT", "/api/v1/routes/"+toString(route.ID), reqBody, userToken)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestUpdateRouteStops(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test data
	stopRepo := repository.NewStopRepository(db)
	routeRepo := repository.NewRouteRepository(db)

	stop1 := &models.Stop{
		Name:      "Stop X",
		Type:      models.StopTypeStop,
		Latitude:  40.7128,
		Longitude: -74.0060,
		Status:    models.StatusActive,
	}
	stopRepo.Create(stop1)

	stop2 := &models.Stop{
		Name:      "Stop Y",
		Type:      models.StopTypeStop,
		Latitude:  40.7580,
		Longitude: -73.9855,
		Status:    models.StatusActive,
	}
	stopRepo.Create(stop2)

	route := &models.Route{
		RouteNumber: "R4",
		Name:        "Route 4",
		Status:      models.StatusActive,
	}
	routeRepo.Create(route)

	// Create admin user
	adminUser := createTestUser(t, db, "admin3@example.com", "admin3", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	t.Run("update route stops as admin", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"stop_ids": []uint{stop1.ID, stop2.ID},
		}

		w := makeRequest(router, "PUT", "/api/v1/routes/"+toString(route.ID)+"/stops", reqBody, adminToken)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestDeleteRoute(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create admin user
	adminUser := createTestUser(t, db, "admin4@example.com", "admin4", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	// Create regular user
	regularUser := createTestUser(t, db, "user3@example.com", "user3", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(regularUser, cfg)
	assert.NoError(t, err)

	t.Run("delete route as admin", func(t *testing.T) {
		// Create route to delete
		routeRepo := repository.NewRouteRepository(db)
		route := &models.Route{
			RouteNumber: "R5",
			Name:        "Route 5",
			Status:      models.StatusActive,
		}
		routeRepo.Create(route)

		w := makeRequest(router, "DELETE", "/api/v1/routes/"+toString(route.ID), nil, adminToken)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("delete route as regular user (should fail)", func(t *testing.T) {
		// Create route to delete
		routeRepo := repository.NewRouteRepository(db)
		route := &models.Route{
			RouteNumber: "R6",
			Name:        "Route 6",
			Status:      models.StatusActive,
		}
		routeRepo.Create(route)

		w := makeRequest(router, "DELETE", "/api/v1/routes/"+toString(route.ID), nil, userToken)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// Helper function to convert uint to string
func toString(n uint) string {
	return fmt.Sprintf("%d", n)
}

