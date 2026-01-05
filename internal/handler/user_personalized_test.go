package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"tj-routes/internal/models"
	"tj-routes/internal/repository"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// =====================
// Helper Functions
// =====================

// uniqueSuffix returns a unique suffix using random number
func uniqueSuffix() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d", rand.Int())
}

// createTestRoute creates a test route
func createTestRoute(t *testing.T, db *gorm.DB, routeNumber, name string) *models.Route {
	routeRepo := repository.NewRouteRepository(db)
	route := &models.Route{
		RouteNumber: routeNumber + uniqueSuffix()[:8],
		Name:        name + uniqueSuffix()[:8],
		Status:      models.StatusActive,
	}
	if err := routeRepo.Create(route); err != nil {
		t.Fatalf("Failed to create test route: %v", err)
	}
	return route
}

// createTestStop creates a test stop
func createTestStop(t *testing.T, db *gorm.DB, name string, lat, lng float64) *models.Stop {
	stopRepo := repository.NewStopRepository(db)
	stop := &models.Stop{
		Name:      name + uniqueSuffix()[:8],
		Type:      models.StopTypeStop,
		Latitude:  lat,
		Longitude: lng,
		Status:    models.StatusActive,
	}
	if err := stopRepo.Create(stop); err != nil {
		t.Fatalf("Failed to create test stop: %v", err)
	}
	return stop
}

// =====================
// Favorite Routes Tests
// =====================

func TestGetFavoriteRoutes(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test user with unique email
	suffix := uniqueSuffix()[:8]
	user := createTestUser(t, db, fmt.Sprintf("fav_routes_%s@example.com", suffix), fmt.Sprintf("fav_routes_%s", suffix), "password123", models.RoleCommonUser)
	token, err := generateTestToken(user, cfg)
	assert.NoError(t, err)

	// Create test routes
	route1 := createTestRoute(t, db, "R001", "Test Route 1")
	route2 := createTestRoute(t, db, "R002", "Test Route 2")

	t.Run("get favorite routes without auth", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/favorites/routes", nil, "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("get empty favorite routes", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/favorites/routes", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		favorites := data["data"].([]interface{})
		assert.Equal(t, 0, len(favorites))
	})

	t.Run("add favorite route", func(t *testing.T) {
		w := makeRequest(router, "POST", fmt.Sprintf("/api/v1/users/me/personalized/favorites/routes/%d", route1.ID), nil, token)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("get favorite routes with data", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/favorites/routes", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		favorites := data["data"].([]interface{})
		assert.Equal(t, 1, len(favorites))
	})

	t.Run("add same route again should return conflict", func(t *testing.T) {
		w := makeRequest(router, "POST", fmt.Sprintf("/api/v1/users/me/personalized/favorites/routes/%d", route1.ID), nil, token)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("check if route is favorite", func(t *testing.T) {
		w := makeRequest(router, "GET", fmt.Sprintf("/api/v1/users/me/personalized/favorites/routes/%d/check", route1.ID), nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		assert.Equal(t, true, data["is_favorite"])
	})

	t.Run("check if non-favorite route is not favorite", func(t *testing.T) {
		w := makeRequest(router, "GET", fmt.Sprintf("/api/v1/users/me/personalized/favorites/routes/%d/check", route2.ID), nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		assert.Equal(t, false, data["is_favorite"])
	})

	t.Run("remove favorite route", func(t *testing.T) {
		w := makeRequest(router, "DELETE", fmt.Sprintf("/api/v1/users/me/personalized/favorites/routes/%d", route1.ID), nil, token)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("get favorite routes after removal", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/favorites/routes", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		favorites := data["data"].([]interface{})
		assert.Equal(t, 0, len(favorites))
	})

	t.Run("pagination test", func(t *testing.T) {
		// Add multiple routes
		route3 := createTestRoute(t, db, "R003", "Test Route 3")
		route4 := createTestRoute(t, db, "R004", "Test Route 4")
		route5 := createTestRoute(t, db, "R005", "Test Route 5")

		makeRequest(router, "POST", fmt.Sprintf("/api/v1/users/me/personalized/favorites/routes/%d", route2.ID), nil, token)
		makeRequest(router, "POST", fmt.Sprintf("/api/v1/users/me/personalized/favorites/routes/%d", route3.ID), nil, token)
		makeRequest(router, "POST", fmt.Sprintf("/api/v1/users/me/personalized/favorites/routes/%d", route4.ID), nil, token)
		makeRequest(router, "POST", fmt.Sprintf("/api/v1/users/me/personalized/favorites/routes/%d", route5.ID), nil, token)

		// Test first page with limit 2
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/favorites/routes?page=1&limit=2", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		favorites := data["data"].([]interface{})
		assert.Equal(t, 2, len(favorites))
		assert.Equal(t, int64(4), int64(data["total"].(float64)))
	})
}

// =====================
// Favorite Stops Tests
// =====================

func TestFavoriteStops(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test user with unique email
	suffix := uniqueSuffix()[:8]
	user := createTestUser(t, db, fmt.Sprintf("fav_stops_%s@example.com", suffix), fmt.Sprintf("fav_stops_%s", suffix), "password123", models.RoleCommonUser)
	token, err := generateTestToken(user, cfg)
	assert.NoError(t, err)

	// Create test stops
	stop1 := createTestStop(t, db, "Test Stop 1", 37.7749, -122.4194)
	stop2 := createTestStop(t, db, "Test Stop 2", 37.7849, -122.4094)

	t.Run("get favorite stops without auth", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/favorites/stops", nil, "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("get empty favorite stops", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/favorites/stops", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		favorites := data["data"].([]interface{})
		assert.Equal(t, 0, len(favorites))
	})

	t.Run("add favorite stop", func(t *testing.T) {
		w := makeRequest(router, "POST", fmt.Sprintf("/api/v1/users/me/personalized/favorites/stops/%d", stop1.ID), nil, token)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Also add stop2
		w = makeRequest(router, "POST", fmt.Sprintf("/api/v1/users/me/personalized/favorites/stops/%d", stop2.ID), nil, token)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("get favorite stops with data", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/favorites/stops", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		favorites := data["data"].([]interface{})
		assert.Equal(t, 2, len(favorites))
	})

	t.Run("check if stop is favorite", func(t *testing.T) {
		w := makeRequest(router, "GET", fmt.Sprintf("/api/v1/users/me/personalized/favorites/stops/%d/check", stop1.ID), nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		assert.Equal(t, true, data["is_favorite"])
	})

	t.Run("remove favorite stop", func(t *testing.T) {
		w := makeRequest(router, "DELETE", fmt.Sprintf("/api/v1/users/me/personalized/favorites/stops/%d", stop1.ID), nil, token)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("get favorite stops after removal", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/favorites/stops", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		favorites := data["data"].([]interface{})
		assert.Equal(t, 1, len(favorites)) // stop2 remains
	})
}

// =====================
// Places Tests
// =====================

func TestPlaces(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test user with unique email
	suffix := uniqueSuffix()[:8]
	user := createTestUser(t, db, fmt.Sprintf("places_%s@example.com", suffix), fmt.Sprintf("places_%s", suffix), "password123", models.RoleCommonUser)
	token, err := generateTestToken(user, cfg)
	assert.NoError(t, err)

	t.Run("get places without auth", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/places", nil, "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("get empty places", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/places", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		places := data["data"].([]interface{})
		assert.Equal(t, 0, len(places))
	})

	t.Run("create place with validation error", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"place_type": "home",
			// Missing name, latitude, longitude
		}
		w := makeRequest(router, "POST", "/api/v1/users/me/personalized/places", reqBody, token)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create place successfully", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"place_type": "home",
			"name":       "My Home",
			"latitude":   37.7749,
			"longitude":  -122.4194,
			"address":    "123 Home St",
			"notes":      "Near the park",
		}
		w := makeRequest(router, "POST", "/api/v1/users/me/personalized/places", reqBody, token)
		assert.Equal(t, http.StatusCreated, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		assert.Equal(t, "home", data["place_type"])
		assert.Equal(t, "My Home", data["name"])
	})

	t.Run("get places after creation", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/places", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		places := data["data"].([]interface{})
		assert.GreaterOrEqual(t, len(places), 1) // At least 1 place exists
	})

	t.Run("create multiple places", func(t *testing.T) {
		// Office
		reqBody := map[string]interface{}{
			"place_type": "office",
			"name":       "Office",
			"latitude":   37.7849,
			"longitude":  -122.4094,
		}
		w := makeRequest(router, "POST", "/api/v1/users/me/personalized/places", reqBody, token)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Gym
		reqBody = map[string]interface{}{
			"place_type": "gym",
			"name":       "Fitness Center",
			"latitude":   37.7649,
			"longitude":  -122.4294,
		}
		w = makeRequest(router, "POST", "/api/v1/users/me/personalized/places", reqBody, token)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("get all places", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/places", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		places := data["data"].([]interface{})
		assert.GreaterOrEqual(t, len(places), 3) // At least 3 places exist
	})

	t.Run("get place by id", func(t *testing.T) {
		// First get all places to get an ID
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/places", nil, token)
		var response Response
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response.Data.(map[string]interface{})
		places := data["data"].([]interface{})
		placeID := uint(places[0].(map[string]interface{})["id"].(float64))

		w = makeRequest(router, "GET", fmt.Sprintf("/api/v1/users/me/personalized/places/%d", placeID), nil, token)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("get non-existent place", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/places/99999", nil, token)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("update place", func(t *testing.T) {
		// First get all places to get an ID
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/places", nil, token)
		var response Response
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response.Data.(map[string]interface{})
		places := data["data"].([]interface{})
		placeID := uint(places[0].(map[string]interface{})["id"].(float64))

		reqBody := map[string]interface{}{
			"name": "Updated Home",
		}
		w = makeRequest(router, "PUT", fmt.Sprintf("/api/v1/users/me/personalized/places/%d", placeID), reqBody, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var updateResponse Response
		json.Unmarshal(w.Body.Bytes(), &updateResponse)
		updateData := updateResponse.Data.(map[string]interface{})
		assert.Equal(t, "Updated Home", updateData["name"])
	})

	t.Run("update non-existent place", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "Updated",
		}
		w := makeRequest(router, "PUT", "/api/v1/users/me/personalized/places/99999", reqBody, token)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("delete place", func(t *testing.T) {
		// First get all places to get an ID
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/places", nil, token)
		var response Response
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response.Data.(map[string]interface{})
		places := data["data"].([]interface{})
		placeID := uint(places[0].(map[string]interface{})["id"].(float64))

		w = makeRequest(router, "DELETE", fmt.Sprintf("/api/v1/users/me/personalized/places/%d", placeID), nil, token)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("delete non-existent place", func(t *testing.T) {
		w := makeRequest(router, "DELETE", "/api/v1/users/me/personalized/places/99999", nil, token)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("get places after deletion", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/places", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		places := data["data"].([]interface{})
		assert.LessOrEqual(t, len(places), 3) // At most 3 places (1 deleted)
	})
}

// =====================
// Recent Views Tests
// =====================

func TestRecentViews(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test user with unique email
	suffix := uniqueSuffix()[:8]
	user := createTestUser(t, db, fmt.Sprintf("recent_%s@example.com", suffix), fmt.Sprintf("recent_%s", suffix), "password123", models.RoleCommonUser)
	token, err := generateTestToken(user, cfg)
	assert.NoError(t, err)

	// Create test route
	route1 := createTestRoute(t, db, "R100", "Recent Route 1")

	// Create test stops
	stop1 := createTestStop(t, db, "Recent Stop 1", 37.7749, -122.4194)
	stop2 := createTestStop(t, db, "Recent Stop 2", 37.7849, -122.4094)

	t.Run("get recent routes without auth", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/recent/routes", nil, "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("get empty recent routes", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/recent/routes", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		favorites := data["data"].([]interface{})
		assert.Equal(t, 0, len(favorites))
	})

	t.Run("record recent navigation", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"view_type":   "navigation",
			"from_stop_id": stop1.ID,
			"to_stop_id":   stop2.ID,
		}
		w := makeRequest(router, "POST", "/api/v1/users/me/personalized/recent/navigations", reqBody, token)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("get recent navigations", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/recent/navigations", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		navigations := data["data"].([]interface{})
		assert.Equal(t, 1, len(navigations))
	})

	t.Run("get recent stops", func(t *testing.T) {
		// Record a stop view
		reqBody := map[string]interface{}{
			"view_type":  "stop",
			"from_stop_id": stop1.ID,
		}
		w := makeRequest(router, "POST", "/api/v1/users/me/personalized/recent/navigations", reqBody, token)
		assert.Equal(t, http.StatusCreated, w.Code)

		w = makeRequest(router, "GET", "/api/v1/users/me/personalized/recent/stops", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		stops := data["data"].([]interface{})
		assert.Equal(t, 1, len(stops))
	})

	t.Run("get recent routes with data", func(t *testing.T) {
		// Record a route view
		reqBody := map[string]interface{}{
			"view_type": "route",
			"route_id":  route1.ID,
		}
		w := makeRequest(router, "POST", "/api/v1/users/me/personalized/recent/navigations", reqBody, token)
		assert.Equal(t, http.StatusCreated, w.Code)

		w = makeRequest(router, "GET", "/api/v1/users/me/personalized/recent/routes", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		routes := data["data"].([]interface{})
		assert.Equal(t, 1, len(routes))
	})

	t.Run("pagination test for recent views", func(t *testing.T) {
		// Create more routes for pagination test
		route2 := createTestRoute(t, db, "R101", "Recent Route 2")
		route3 := createTestRoute(t, db, "R102", "Recent Route 3")

		// Record multiple route views
		for i := 0; i < 5; i++ {
			reqBody := map[string]interface{}{
				"view_type": "route",
				"route_id":  route2.ID,
			}
			makeRequest(router, "POST", "/api/v1/users/me/personalized/recent/navigations", reqBody, token)
		}

		for i := 0; i < 5; i++ {
			reqBody := map[string]interface{}{
				"view_type": "route",
				"route_id":  route3.ID,
			}
			makeRequest(router, "POST", "/api/v1/users/me/personalized/recent/navigations", reqBody, token)
		}

		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/recent/routes?page=1&limit=2", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		routes := data["data"].([]interface{})
		assert.Equal(t, 2, len(routes))
	})
}

// =====================
// Saved Navigations Tests
// =====================

func TestSavedNavigations(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test user with unique email
	suffix := uniqueSuffix()[:8]
	user := createTestUser(t, db, fmt.Sprintf("saved_nav_%s@example.com", suffix), fmt.Sprintf("saved_nav_%s", suffix), "password123", models.RoleCommonUser)
	token, err := generateTestToken(user, cfg)
	assert.NoError(t, err)

	// Create test stops
	stop1 := createTestStop(t, db, "Saved Nav Stop 1", 37.7749, -122.4194)
	stop2 := createTestStop(t, db, "Saved Nav Stop 2", 37.7849, -122.4094)
	stop3 := createTestStop(t, db, "Saved Nav Stop 3", 37.7949, -122.3994)

	t.Run("get saved navigations without auth", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/saved-navigations", nil, "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("get empty saved navigations", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/saved-navigations", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		navigations := data["data"].([]interface{})
		assert.Equal(t, 0, len(navigations))
	})

	t.Run("create saved navigation with invalid navigation point", func(t *testing.T) {
		// Both from and to are empty
		reqBody := map[string]interface{}{
			"name": "Test Navigation",
		}
		w := makeRequest(router, "POST", "/api/v1/users/me/personalized/saved-navigations", reqBody, token)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create saved navigation successfully", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":         "Morning Commute",
			"from_stop_id": stop1.ID,
			"to_stop_id":   stop2.ID,
		}
		w := makeRequest(router, "POST", "/api/v1/users/me/personalized/saved-navigations", reqBody, token)
		assert.Equal(t, http.StatusCreated, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		assert.Equal(t, "Morning Commute", data["name"])
	})

	t.Run("get saved navigations with data", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/saved-navigations", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		navigations := data["data"].([]interface{})
		assert.Equal(t, 1, len(navigations))
	})

	t.Run("create more saved navigations", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":         "Evening Commute",
			"from_stop_id": stop2.ID,
			"to_stop_id":   stop1.ID,
		}
		w := makeRequest(router, "POST", "/api/v1/users/me/personalized/saved-navigations", reqBody, token)
		assert.Equal(t, http.StatusCreated, w.Code)

		reqBody = map[string]interface{}{
			"name":         "Weekend Trip",
			"from_stop_id": stop1.ID,
			"to_stop_id":   stop3.ID,
		}
		w = makeRequest(router, "POST", "/api/v1/users/me/personalized/saved-navigations", reqBody, token)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("pagination test", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/saved-navigations?page=1&limit=2", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		navigations := data["data"].([]interface{})
		assert.Equal(t, 2, len(navigations))
		assert.Equal(t, int64(3), int64(data["total"].(float64)))
	})

	t.Run("update saved navigation", func(t *testing.T) {
		// Get first navigation ID
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/saved-navigations?page=1&limit=1", nil, token)
		var response Response
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response.Data.(map[string]interface{})
		navigations := data["data"].([]interface{})
		navID := uint(navigations[0].(map[string]interface{})["id"].(float64))

		reqBody := map[string]interface{}{
			"name": "Updated Morning Commute",
		}
		w = makeRequest(router, "PUT", fmt.Sprintf("/api/v1/users/me/personalized/saved-navigations/%d", navID), reqBody, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var updateResponse Response
		json.Unmarshal(w.Body.Bytes(), &updateResponse)
		updateData := updateResponse.Data.(map[string]interface{})
		assert.Equal(t, "Updated Morning Commute", updateData["name"])
	})

	t.Run("update non-existent navigation", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "Updated",
		}
		w := makeRequest(router, "PUT", "/api/v1/users/me/personalized/saved-navigations/99999", reqBody, token)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("delete saved navigation", func(t *testing.T) {
		// Get first navigation ID
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/saved-navigations?page=1&limit=1", nil, token)
		var response Response
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response.Data.(map[string]interface{})
		navigations := data["data"].([]interface{})
		navID := uint(navigations[0].(map[string]interface{})["id"].(float64))

		w = makeRequest(router, "DELETE", fmt.Sprintf("/api/v1/users/me/personalized/saved-navigations/%d", navID), nil, token)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("delete non-existent navigation", func(t *testing.T) {
		w := makeRequest(router, "DELETE", "/api/v1/users/me/personalized/saved-navigations/99999", nil, token)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("get saved navigations after deletion", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/saved-navigations", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		navigations := data["data"].([]interface{})
		assert.Equal(t, 2, len(navigations))
	})
}

// =====================
// Analytics Tests
// =====================

func TestAnalytics(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test user with unique email
	suffix := uniqueSuffix()[:8]
	user := createTestUser(t, db, fmt.Sprintf("analytics_%s@example.com", suffix), fmt.Sprintf("analytics_%s", suffix), "password123", models.RoleCommonUser)
	token, err := generateTestToken(user, cfg)
	assert.NoError(t, err)

	t.Run("get analytics without auth", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/analytics", nil, "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("get empty analytics", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/analytics", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		// All counts should be zero for new user
		assert.Equal(t, float64(0), data["total_check_ins"])
		assert.Equal(t, float64(0), data["total_routes_traveled"])
		assert.Equal(t, float64(0), data["total_unique_routes"])
		assert.Equal(t, float64(0), data["total_duration_seconds"])
		assert.Equal(t, float64(0), data["favorite_places_count"])
		assert.Equal(t, float64(0), data["favorite_routes_count"])
		assert.Equal(t, float64(0), data["favorite_stops_count"])
		assert.Equal(t, float64(0), data["saved_navigations_count"])
	})

	t.Run("analytics reflect favorite data", func(t *testing.T) {
		// Create and add a favorite route
		route1 := createTestRoute(t, db, "R200", "Analytics Route 1")
		makeRequest(router, "POST", fmt.Sprintf("/api/v1/users/me/personalized/favorites/routes/%d", route1.ID), nil, token)

		// Create and add a favorite stop
		stop1 := createTestStop(t, db, "Analytics Stop 1", 37.7749, -122.4194)
		makeRequest(router, "POST", fmt.Sprintf("/api/v1/users/me/personalized/favorites/stops/%d", stop1.ID), nil, token)

		// Create a place
		reqBody := map[string]interface{}{
			"place_type": "home",
			"name":       "My Home",
			"latitude":   37.7749,
			"longitude":  -122.4194,
		}
		makeRequest(router, "POST", "/api/v1/users/me/personalized/places", reqBody, token)

		// Create a saved navigation
		reqBody = map[string]interface{}{
			"name":         "Test Nav",
			"from_stop_id": stop1.ID,
			"to_stop_id":   stop1.ID,
		}
		makeRequest(router, "POST", "/api/v1/users/me/personalized/saved-navigations", reqBody, token)

		// Get analytics
		w := makeRequest(router, "GET", "/api/v1/users/me/personalized/analytics", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		assert.Equal(t, float64(1), data["favorite_routes_count"])
		assert.Equal(t, float64(1), data["favorite_stops_count"])
		assert.Equal(t, float64(1), data["favorite_places_count"])
		assert.Equal(t, float64(1), data["saved_navigations_count"])
	})
}

// =====================
// Authorization Tests
// =====================

func TestAuthorization(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create two users with unique emails
	suffix := uniqueSuffix()[:8]
	user1 := createTestUser(t, db, fmt.Sprintf("auth1_%s@example.com", suffix), fmt.Sprintf("auth1_%s", suffix), "password123", models.RoleCommonUser)
	token1, err := generateTestToken(user1, cfg)
	assert.NoError(t, err)

	user2 := createTestUser(t, db, fmt.Sprintf("auth2_%s@example.com", suffix), fmt.Sprintf("auth2_%s", suffix), "password123", models.RoleCommonUser)
	token2, err := generateTestToken(user2, cfg)
	assert.NoError(t, err)

	// User1 creates a place
	placeReq := map[string]interface{}{
		"place_type": "home",
		"name":       "User1's Home",
		"latitude":   37.7749,
		"longitude":  -122.4194,
	}
	w := makeRequest(router, "POST", "/api/v1/users/me/personalized/places", placeReq, token1)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createResponse Response
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	createData := createResponse.Data.(map[string]interface{})
	placeID := uint(createData["id"].(float64))

	t.Run("user cannot access another user's place", func(t *testing.T) {
		w := makeRequest(router, "GET", fmt.Sprintf("/api/v1/users/me/personalized/places/%d", placeID), nil, token2)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("user cannot update another user's place", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "Hacked!",
		}
		w := makeRequest(router, "PUT", fmt.Sprintf("/api/v1/users/me/personalized/places/%d", placeID), reqBody, token2)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("user cannot delete another user's place", func(t *testing.T) {
		w := makeRequest(router, "DELETE", fmt.Sprintf("/api/v1/users/me/personalized/places/%d", placeID), nil, token2)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("user can access their own place", func(t *testing.T) {
		w := makeRequest(router, "GET", fmt.Sprintf("/api/v1/users/me/personalized/places/%d", placeID), nil, token1)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
