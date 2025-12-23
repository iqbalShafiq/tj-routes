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

func TestListVehicles(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test data
	routeRepo := repository.NewRouteRepository(db)
	route := &models.Route{
		RouteNumber: "RV1",
		Name:        "Route V1",
		Status:      models.StatusActive,
	}
	routeRepo.Create(route)

	vehicleRepo := repository.NewVehicleRepository(db)
	vehicle1 := &models.Vehicle{
		VehiclePlate: "ABC-123",
		RouteID:      route.ID,
		VehicleType:  "bus",
		Capacity:     50,
		Status:       models.StatusActive,
	}
	vehicleRepo.Create(vehicle1)

	vehicle2 := &models.Vehicle{
		VehiclePlate: "XYZ-789",
		RouteID:      route.ID,
		VehicleType:  "minibus",
		Capacity:     25,
		Status:       models.StatusActive,
	}
	vehicleRepo.Create(vehicle2)

	t.Run("list vehicles without auth", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/vehicles", nil, "")

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		vehicles := data["vehicles"].([]interface{})
		assert.GreaterOrEqual(t, len(vehicles), 2)
	})

	t.Run("list vehicles with pagination", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/vehicles?page=1&limit=5", nil, "")

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("list vehicles with route_id filter", func(t *testing.T) {
		w := makeRequest(router, "GET", fmt.Sprintf("/api/v1/vehicles?route_id=%d", route.ID), nil, "")

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestGetVehicle(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test data
	routeRepo := repository.NewRouteRepository(db)
	route := &models.Route{
		RouteNumber: "RV2",
		Name:        "Route V2",
		Status:      models.StatusActive,
	}
	routeRepo.Create(route)

	vehicleRepo := repository.NewVehicleRepository(db)
	vehicle := &models.Vehicle{
		VehiclePlate: "TEST-001",
		RouteID:      route.ID,
		VehicleType:  "bus",
		Capacity:     50,
		Status:       models.StatusActive,
	}
	vehicleRepo.Create(vehicle)

	t.Run("get vehicle without auth", func(t *testing.T) {
		w := makeRequest(router, "GET", fmt.Sprintf("/api/v1/vehicles/%d", vehicle.ID), nil, "")

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("get non-existent vehicle", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/vehicles/99999", nil, "")

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestCreateVehicle(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test route
	routeRepo := repository.NewRouteRepository(db)
	route := &models.Route{
		RouteNumber: "RV3",
		Name:        "Route V3",
		Status:      models.StatusActive,
	}
	routeRepo.Create(route)

	// Create admin user
	adminUser := createTestUser(t, db, "admin8@example.com", "admin8", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	// Create regular user
	regularUser := createTestUser(t, db, "user7@example.com", "user7", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(regularUser, cfg)
	assert.NoError(t, err)

	t.Run("create vehicle as admin", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"vehicle_plate": "NEW-001",
			"route_id":      route.ID,
			"vehicle_type":  "bus",
			"capacity":      50,
			"status":        "active",
		}

		w := makeRequest(router, "POST", "/api/v1/vehicles", reqBody, adminToken)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("create vehicle as regular user (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"vehicle_plate": "UNAUTH-001",
			"route_id":      route.ID,
			"vehicle_type":  "bus",
		}

		w := makeRequest(router, "POST", "/api/v1/vehicles", reqBody, userToken)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("create vehicle without auth (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"vehicle_plate": "PUBLIC-001",
			"route_id":      route.ID,
		}

		w := makeRequest(router, "POST", "/api/v1/vehicles", reqBody, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("create vehicle with invalid route_id", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"vehicle_plate": "INVALID-001",
			"route_id":      99999, // Non-existent route
		}

		w := makeRequest(router, "POST", "/api/v1/vehicles", reqBody, adminToken)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateVehicle(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test data
	routeRepo := repository.NewRouteRepository(db)
	route := &models.Route{
		RouteNumber: "RV4",
		Name:        "Route V4",
		Status:      models.StatusActive,
	}
	routeRepo.Create(route)

	vehicleRepo := repository.NewVehicleRepository(db)
	vehicle := &models.Vehicle{
		VehiclePlate: "UPDATE-001",
		RouteID:      route.ID,
		VehicleType:  "bus",
		Capacity:     50,
		Status:       models.StatusActive,
	}
	vehicleRepo.Create(vehicle)

	// Create admin user
	adminUser := createTestUser(t, db, "admin9@example.com", "admin9", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	// Create regular user
	regularUser := createTestUser(t, db, "user8@example.com", "user8", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(regularUser, cfg)
	assert.NoError(t, err)

	t.Run("update vehicle as admin", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"vehicle_type": "minibus",
			"capacity":     30,
		}

		w := makeRequest(router, "PUT", fmt.Sprintf("/api/v1/vehicles/%d", vehicle.ID), reqBody, adminToken)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("update vehicle as regular user (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"capacity": 40,
		}

		w := makeRequest(router, "PUT", fmt.Sprintf("/api/v1/vehicles/%d", vehicle.ID), reqBody, userToken)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestDeleteVehicle(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create admin user
	adminUser := createTestUser(t, db, "admin10@example.com", "admin10", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	// Create regular user
	regularUser := createTestUser(t, db, "user9@example.com", "user9", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(regularUser, cfg)
	assert.NoError(t, err)

	t.Run("delete vehicle as admin", func(t *testing.T) {
		// Create test data
		routeRepo := repository.NewRouteRepository(db)
		route := &models.Route{
			RouteNumber: "RV5",
			Name:        "Route V5",
			Status:      models.StatusActive,
		}
		routeRepo.Create(route)

		vehicleRepo := repository.NewVehicleRepository(db)
		vehicle := &models.Vehicle{
			VehiclePlate: "DELETE-001",
			RouteID:      route.ID,
			VehicleType:  "bus",
			Capacity:     50,
			Status:       models.StatusActive,
		}
		vehicleRepo.Create(vehicle)

		w := makeRequest(router, "DELETE", fmt.Sprintf("/api/v1/vehicles/%d", vehicle.ID), nil, adminToken)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("delete vehicle as regular user (should fail)", func(t *testing.T) {
		// Create test data
		routeRepo := repository.NewRouteRepository(db)
		route := &models.Route{
			RouteNumber: "RV6",
			Name:        "Route V6",
			Status:      models.StatusActive,
		}
		routeRepo.Create(route)

		vehicleRepo := repository.NewVehicleRepository(db)
		vehicle := &models.Vehicle{
			VehiclePlate: "DELETE-002",
			RouteID:      route.ID,
			VehicleType:  "bus",
			Capacity:     50,
			Status:       models.StatusActive,
		}
		vehicleRepo.Create(vehicle)

		w := makeRequest(router, "DELETE", fmt.Sprintf("/api/v1/vehicles/%d", vehicle.ID), nil, userToken)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

