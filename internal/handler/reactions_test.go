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

func TestReactToReport(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test user
	user := createTestUser(t, db, "reactor1@example.com", "reactor1", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(user, cfg)
	assert.NoError(t, err)

	// Create report
	reportRepo := repository.NewReportRepository(db)
	report := &models.Report{
		UserID:      user.ID,
		Type:        models.ReportTypeRouteIssue,
		Title:       "Report to React",
		Description: "Description",
		Status:      models.ReportStatusPending,
	}
	reportRepo.Create(report)

	t.Run("upvote report", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"type": "upvote",
		}

		w := makeRequest(router, "POST", fmt.Sprintf("/api/v1/reports/%d/react", report.ID), reqBody, userToken)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("downvote report", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"type": "downvote",
		}

		w := makeRequest(router, "POST", fmt.Sprintf("/api/v1/reports/%d/react", report.ID), reqBody, userToken)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("react without auth (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"type": "upvote",
		}

		w := makeRequest(router, "POST", fmt.Sprintf("/api/v1/reports/%d/react", report.ID), reqBody, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("react with invalid type", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"type": "invalid",
		}

		w := makeRequest(router, "POST", fmt.Sprintf("/api/v1/reports/%d/react", report.ID), reqBody, userToken)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestRemoveReactionFromReport(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test user
	user := createTestUser(t, db, "reactor2@example.com", "reactor2", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(user, cfg)
	assert.NoError(t, err)

	// Create report
	reportRepo := repository.NewReportRepository(db)
	report := &models.Report{
		UserID:      user.ID,
		Type:        models.ReportTypeRouteIssue,
		Title:       "Report to Unreact",
		Description: "Description",
		Status:      models.ReportStatusPending,
	}
	reportRepo.Create(report)

	t.Run("remove reaction from report", func(t *testing.T) {
		// First add a reaction
		reqBody := map[string]interface{}{
			"type": "upvote",
		}
		w := makeRequest(router, "POST", fmt.Sprintf("/api/v1/reports/%d/react", report.ID), reqBody, userToken)
		assert.Equal(t, http.StatusOK, w.Code)

		// Then remove it
		w = makeRequest(router, "DELETE", fmt.Sprintf("/api/v1/reports/%d/react", report.ID), nil, userToken)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("remove reaction without auth (should fail)", func(t *testing.T) {
		w := makeRequest(router, "DELETE", fmt.Sprintf("/api/v1/reports/%d/react", report.ID), nil, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestReactToComment(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test user
	user := createTestUser(t, db, "reactor3@example.com", "reactor3", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(user, cfg)
	assert.NoError(t, err)

	// Create report and comment
	reportRepo := repository.NewReportRepository(db)
	report := &models.Report{
		UserID:      user.ID,
		Type:        models.ReportTypeRouteIssue,
		Title:       "Report for Comment Reaction",
		Description: "Description",
		Status:      models.ReportStatusPending,
	}
	reportRepo.Create(report)

	commentRepo := repository.NewCommentRepository(db)
	comment := &models.Comment{
		ReportID: &report.ID,
		UserID:   user.ID,
		Content:  "Comment to react",
	}
	commentRepo.Create(comment)

	t.Run("upvote comment", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"type": "upvote",
		}

		w := makeRequest(router, "POST", fmt.Sprintf("/api/v1/comments/%d/react", comment.ID), reqBody, userToken)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("downvote comment", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"type": "downvote",
		}

		w := makeRequest(router, "POST", fmt.Sprintf("/api/v1/comments/%d/react", comment.ID), reqBody, userToken)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("react without auth (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"type": "upvote",
		}

		w := makeRequest(router, "POST", fmt.Sprintf("/api/v1/comments/%d/react", comment.ID), reqBody, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRemoveReactionFromComment(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test user
	user := createTestUser(t, db, "reactor4@example.com", "reactor4", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(user, cfg)
	assert.NoError(t, err)

	// Create report and comment
	reportRepo := repository.NewReportRepository(db)
	report := &models.Report{
		UserID:      user.ID,
		Type:        models.ReportTypeRouteIssue,
		Title:       "Report for Comment Unreact",
		Description: "Description",
		Status:      models.ReportStatusPending,
	}
	reportRepo.Create(report)

	commentRepo := repository.NewCommentRepository(db)
	comment := &models.Comment{
		ReportID: &report.ID,
		UserID:   user.ID,
		Content:  "Comment to unreact",
	}
	commentRepo.Create(comment)

	t.Run("remove reaction from comment", func(t *testing.T) {
		// First add a reaction
		reqBody := map[string]interface{}{
			"type": "upvote",
		}
		w := makeRequest(router, "POST", fmt.Sprintf("/api/v1/comments/%d/react", comment.ID), reqBody, userToken)
		assert.Equal(t, http.StatusOK, w.Code)

		// Then remove it
		w = makeRequest(router, "DELETE", fmt.Sprintf("/api/v1/comments/%d/react", comment.ID), nil, userToken)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("remove reaction without auth (should fail)", func(t *testing.T) {
		w := makeRequest(router, "DELETE", fmt.Sprintf("/api/v1/comments/%d/react", comment.ID), nil, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

