package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tj-routes/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	t.Run("successful registration", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email":    "test@example.com",
			"username": "testuser",
			"password": "password123",
		}

		w := makeRequest(router, "POST", "/api/v1/auth/register", reqBody, "")

		assert.Equal(t, http.StatusCreated, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)
	})

	t.Run("duplicate email", func(t *testing.T) {
		// Create user first
		createTestUser(t, db, "duplicate@example.com", "existinguser", "password123", models.RoleCommonUser)

		reqBody := map[string]interface{}{
			"email":    "duplicate@example.com",
			"username": "newuser",
			"password": "password123",
		}

		w := makeRequest(router, "POST", "/api/v1/auth/register", reqBody, "")

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid email format", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email":    "invalid-email",
			"username": "testuser",
			"password": "password123",
		}

		w := makeRequest(router, "POST", "/api/v1/auth/register", reqBody, "")

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("password too short", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email":    "test2@example.com",
			"username": "testuser",
			"password": "12345", // Less than 6 characters
		}

		w := makeRequest(router, "POST", "/api/v1/auth/register", reqBody, "")

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("username too short", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email":    "test3@example.com",
			"username": "ab", // Less than 3 characters
			"password": "password123",
		}

		w := makeRequest(router, "POST", "/api/v1/auth/register", reqBody, "")

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestLogin(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test user
	createTestUser(t, db, "login@example.com", "loginuser", "password123", models.RoleCommonUser)

	t.Run("successful login", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email":    "login@example.com",
			"password": "password123",
		}

		w := makeRequest(router, "POST", "/api/v1/auth/login", reqBody, "")

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)

		// Check that tokens are present
		data := response.Data.(map[string]interface{})
		assert.NotEmpty(t, data["access_token"])
		assert.NotEmpty(t, data["refresh_token"])
		assert.NotNil(t, data["user"])
	})

	t.Run("invalid email", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email":    "nonexistent@example.com",
			"password": "password123",
		}

		w := makeRequest(router, "POST", "/api/v1/auth/login", reqBody, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid password", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email":    "login@example.com",
			"password": "wrongpassword",
		}

		w := makeRequest(router, "POST", "/api/v1/auth/login", reqBody, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("missing email", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"password": "password123",
		}

		w := makeRequest(router, "POST", "/api/v1/auth/login", reqBody, "")

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestOAuthInitiate(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	t.Run("unsupported provider", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/auth/oauth/github", nil, "")

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("google provider redirects", func(t *testing.T) {
		// Note: This test checks that the endpoint returns a redirect
		// The actual OAuth flow would require external OAuth provider
		req := httptest.NewRequest("GET", "/api/v1/auth/oauth/google", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should redirect to Google OAuth
		assert.True(t, w.Code == http.StatusTemporaryRedirect || w.Code == http.StatusFound)
	})
}

