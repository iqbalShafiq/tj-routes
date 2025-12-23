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

func TestCreateReport(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test route
	routeRepo := repository.NewRouteRepository(db)
	route := &models.Route{
		RouteNumber: "RR1",
		Name:        "Route R1",
		Status:      models.StatusActive,
	}
	routeRepo.Create(route)

	// Create authenticated user
	user := createTestUser(t, db, "reporter@example.com", "reporter", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(user, cfg)
	assert.NoError(t, err)

	t.Run("create report as guest (without auth)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"type":        "route_issue",
			"title":       "Test Report",
			"description": "This is a test report",
		}

		w := makeRequest(router, "POST", "/api/v1/reports", reqBody, "")

		assert.Equal(t, http.StatusCreated, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("create report as authenticated user", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"type":           "route_issue",
			"title":          "User Report",
			"description":    "Report from authenticated user",
			"related_route_id": route.ID,
		}

		w := makeRequest(router, "POST", "/api/v1/reports", reqBody, userToken)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("create report with missing required fields", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"title": "Incomplete Report",
		}

		w := makeRequest(router, "POST", "/api/v1/reports", reqBody, "")

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestListReports(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test users
	user1 := createTestUser(t, db, "user1@example.com", "user1", "password123", models.RoleCommonUser)
	user1Token, err := generateTestToken(user1, cfg)
	assert.NoError(t, err)

	user2 := createTestUser(t, db, "user2@example.com", "user2", "password123", models.RoleCommonUser)
	_, err = generateTestToken(user2, cfg)
	assert.NoError(t, err)

	adminUser := createTestUser(t, db, "admin11@example.com", "admin11", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	// Create reports for user1
	reportRepo := repository.NewReportRepository(db)
	report1 := &models.Report{
		UserID:      user1.ID,
		Type:        models.ReportTypeRouteIssue,
		Title:       "User1 Report 1",
		Description: "Description 1",
		Status:      models.ReportStatusPending,
	}
	reportRepo.Create(report1)

	report2 := &models.Report{
		UserID:      user1.ID,
		Type:        models.ReportTypeStopIssue,
		Title:       "User1 Report 2",
		Description: "Description 2",
		Status:      models.ReportStatusPending,
	}
	reportRepo.Create(report2)

	// Create report for user2
	report3 := &models.Report{
		UserID:      user2.ID,
		Type:        models.ReportTypeRouteIssue,
		Title:       "User2 Report",
		Description: "Description 3",
		Status:      models.ReportStatusPending,
	}
	reportRepo.Create(report3)

	t.Run("list own reports as regular user", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/reports", nil, user1Token)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		reports := data["reports"].([]interface{})
		// Should see only user1's reports (2 reports)
		assert.Equal(t, 2, len(reports))
	})

	t.Run("list all reports as admin", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/reports", nil, adminToken)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		reports := data["reports"].([]interface{})
		// Admin should see all reports (at least 3)
		assert.GreaterOrEqual(t, len(reports), 3)
	})

	t.Run("list reports without auth (should fail)", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/v1/reports", nil, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestGetReport(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test users
	user1 := createTestUser(t, db, "user3@example.com", "user3", "password123", models.RoleCommonUser)
	user1Token, err := generateTestToken(user1, cfg)
	assert.NoError(t, err)

	user2 := createTestUser(t, db, "user4@example.com", "user4", "password123", models.RoleCommonUser)
	user2Token, err := generateTestToken(user2, cfg)
	assert.NoError(t, err)

	adminUser := createTestUser(t, db, "admin12@example.com", "admin12", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	// Create report for user1
	reportRepo := repository.NewReportRepository(db)
	report := &models.Report{
		UserID:      user1.ID,
		Type:        models.ReportTypeRouteIssue,
		Title:       "Private Report",
		Description: "This is user1's report",
		Status:      models.ReportStatusPending,
	}
	reportRepo.Create(report)

	t.Run("get own report as regular user", func(t *testing.T) {
		w := makeRequest(router, "GET", fmt.Sprintf("/api/v1/reports/%d", report.ID), nil, user1Token)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("get another user's report as regular user (should fail)", func(t *testing.T) {
		w := makeRequest(router, "GET", fmt.Sprintf("/api/v1/reports/%d", report.ID), nil, user2Token)

		// Should fail or return not found depending on implementation
		assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusForbidden)
	})

	t.Run("get any report as admin", func(t *testing.T) {
		w := makeRequest(router, "GET", fmt.Sprintf("/api/v1/reports/%d", report.ID), nil, adminToken)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})
}

func TestUpdateReportStatus(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test users
	user := createTestUser(t, db, "user5@example.com", "user5", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(user, cfg)
	assert.NoError(t, err)

	adminUser := createTestUser(t, db, "admin13@example.com", "admin13", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	// Create report
	reportRepo := repository.NewReportRepository(db)
	report := &models.Report{
		UserID:      user.ID,
		Type:        models.ReportTypeRouteIssue,
		Title:       "Report to Update",
		Description: "Description",
		Status:      models.ReportStatusPending,
	}
	reportRepo.Create(report)

	t.Run("update report status as admin", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"status":      "reviewed",
			"admin_notes": "Reviewed by admin",
		}

		w := makeRequest(router, "PUT", fmt.Sprintf("/api/v1/reports/%d/status", report.ID), reqBody, adminToken)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("update report status as regular user (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"status": "resolved",
		}

		w := makeRequest(router, "PUT", fmt.Sprintf("/api/v1/reports/%d/status", report.ID), reqBody, userToken)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestDeleteReport(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test users
	user := createTestUser(t, db, "user6@example.com", "user6", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(user, cfg)
	assert.NoError(t, err)

	adminUser := createTestUser(t, db, "admin14@example.com", "admin14", "password123", models.RoleAdmin)
	adminToken, err := generateTestToken(adminUser, cfg)
	assert.NoError(t, err)

	t.Run("delete report as admin", func(t *testing.T) {
		// Create report to delete
		reportRepo := repository.NewReportRepository(db)
		report := &models.Report{
			UserID:      user.ID,
			Type:        models.ReportTypeRouteIssue,
			Title:       "Report to Delete",
			Description: "Description",
			Status:      models.ReportStatusPending,
		}
		reportRepo.Create(report)

		w := makeRequest(router, "DELETE", fmt.Sprintf("/api/v1/reports/%d", report.ID), nil, adminToken)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("delete report as regular user (should fail)", func(t *testing.T) {
		// Create report to delete
		reportRepo := repository.NewReportRepository(db)
		report := &models.Report{
			UserID:      user.ID,
			Type:        models.ReportTypeRouteIssue,
			Title:       "Report to Delete 2",
			Description: "Description",
			Status:      models.ReportStatusPending,
		}
		reportRepo.Create(report)

		w := makeRequest(router, "DELETE", fmt.Sprintf("/api/v1/reports/%d", report.ID), nil, userToken)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

