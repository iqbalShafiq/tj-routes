package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"tj-routes/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestGetUserProfile(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test users
	user1 := createTestUser(t, db, "profile1@example.com", "profile1", "password123", models.RoleCommonUser)
	user1Token, err := generateTestToken(user1, cfg)
	assert.NoError(t, err)

	user2 := createTestUser(t, db, "profile2@example.com", "profile2", "password123", models.RoleCommonUser)
	_, err = generateTestToken(user2, cfg)
	assert.NoError(t, err)

	t.Run("get own profile", func(t *testing.T) {
		w := makeRequest(router, "GET", fmt.Sprintf("/api/v1/users/%d/profile", user1.ID), nil, user1Token)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		assert.NotNil(t, data["user"])
	})

	t.Run("get another user's profile", func(t *testing.T) {
		w := makeRequest(router, "GET", fmt.Sprintf("/api/v1/users/%d/profile", user2.ID), nil, user1Token)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("get profile without auth (should fail)", func(t *testing.T) {
		w := makeRequest(router, "GET", fmt.Sprintf("/api/v1/users/%d/profile", user1.ID), nil, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("get non-existent user profile", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/99999/profile", nil, user1Token)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestListUsers(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create admin user
	adminUser := createTestUser(t, db, "admin15@example.com", "admin15", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	// Create regular user
	regularUser := createTestUser(t, db, "user10@example.com", "user10", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(regularUser, cfg)
	assert.NoError(t, err)

	// Create additional users
	createTestUser(t, db, "user11@example.com", "user11", "password123", models.RoleCommonUser)
	createTestUser(t, db, "user12@example.com", "user12", "password123", models.RoleCommonUser)

	t.Run("list users as admin", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users", nil, adminToken)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		users := data["users"].([]interface{})
		// Should see at least 4 users (admin15, user10, user11, user12, plus system user)
		assert.GreaterOrEqual(t, len(users), 4)
	})

	t.Run("list users as regular user (should fail)", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users", nil, userToken)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("list users without auth (should fail)", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users", nil, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestGetUser(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create admin user
	adminUser := createTestUser(t, db, "admin16@example.com", "admin16", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	// Create regular user
	regularUser := createTestUser(t, db, "user13@example.com", "user13", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(regularUser, cfg)
	assert.NoError(t, err)

	targetUser := createTestUser(t, db, "target@example.com", "target", "password123", models.RoleCommonUser)

	t.Run("get user as admin", func(t *testing.T) {
		w := makeRequest(router, "GET", fmt.Sprintf("/api/v1/users/%d", targetUser.ID), nil, adminToken)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("get user as regular user (should fail)", func(t *testing.T) {
		w := makeRequest(router, "GET", fmt.Sprintf("/api/v1/users/%d", targetUser.ID), nil, userToken)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("get non-existent user", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/users/99999", nil, adminToken)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestUpdateUserRole(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create admin user
	adminUser := createTestUser(t, db, "admin17@example.com", "admin17", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	// Create regular user
	regularUser := createTestUser(t, db, "user14@example.com", "user14", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(regularUser, cfg)
	assert.NoError(t, err)

	targetUser := createTestUser(t, db, "target2@example.com", "target2", "password123", models.RoleCommonUser)

	t.Run("update user role as admin", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"role": "admin",
		}

		w := makeRequest(router, "PUT", fmt.Sprintf("/api/v1/users/%d/role", targetUser.ID), reqBody, adminToken)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("update user role as regular user (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"role": "admin",
		}

		w := makeRequest(router, "PUT", fmt.Sprintf("/api/v1/users/%d/role", targetUser.ID), reqBody, userToken)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("update user role without auth (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"role": "admin",
		}

		w := makeRequest(router, "PUT", fmt.Sprintf("/api/v1/users/%d/role", targetUser.ID), reqBody, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

