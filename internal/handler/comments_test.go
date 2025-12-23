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

func TestGetComments(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test users
	user := createTestUser(t, db, "commenter1@example.com", "commenter1", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(user, cfg)
	assert.NoError(t, err)

	// Create report
	reportRepo := repository.NewReportRepository(db)
	report := &models.Report{
		UserID:      user.ID,
		Type:        models.ReportTypeRouteIssue,
		Title:       "Report with Comments",
		Description: "Description",
		Status:      models.ReportStatusPending,
	}
	reportRepo.Create(report)

	// Create comments
	commentRepo := repository.NewCommentRepository(db)
	comment1 := &models.Comment{
		ReportID: report.ID,
		UserID:   user.ID,
		Content:  "First comment",
	}
	commentRepo.Create(comment1)

	comment2 := &models.Comment{
		ReportID: report.ID,
		UserID:   user.ID,
		Content:  "Second comment",
	}
	commentRepo.Create(comment2)

	t.Run("get comments for report", func(t *testing.T) {
		w := makeRequest(router, "GET", fmt.Sprintf("/api/v1/reports/%d/comments", report.ID), nil, userToken)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		comments := data["comments"].([]interface{})
		assert.GreaterOrEqual(t, len(comments), 2)
	})
}

func TestCreateComment(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test users
	user := createTestUser(t, db, "commenter2@example.com", "commenter2", "password123", models.RoleCommonUser)
	userToken, err := generateTestToken(user, cfg)
	assert.NoError(t, err)

	// Create report
	reportRepo := repository.NewReportRepository(db)
	report := &models.Report{
		UserID:      user.ID,
		Type:        models.ReportTypeRouteIssue,
		Title:       "Report for Comment",
		Description: "Description",
		Status:      models.ReportStatusPending,
	}
	reportRepo.Create(report)

	t.Run("create comment", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"content": "This is a test comment",
		}

		w := makeRequest(router, "POST", fmt.Sprintf("/api/v1/reports/%d/comments", report.ID), reqBody, userToken)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("create comment without auth (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"content": "Unauthorized comment",
		}

		w := makeRequest(router, "POST", fmt.Sprintf("/api/v1/reports/%d/comments", report.ID), reqBody, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("create comment with missing content", func(t *testing.T) {
		reqBody := map[string]interface{}{}

		w := makeRequest(router, "POST", fmt.Sprintf("/api/v1/reports/%d/comments", report.ID), reqBody, userToken)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create reply comment", func(t *testing.T) {
		// Create parent comment first
		commentRepo := repository.NewCommentRepository(db)
		parentComment := &models.Comment{
			ReportID: report.ID,
			UserID:   user.ID,
			Content:  "Parent comment",
		}
		commentRepo.Create(parentComment)

		reqBody := map[string]interface{}{
			"content":   "This is a reply",
			"parent_id": parentComment.ID,
		}

		w := makeRequest(router, "POST", fmt.Sprintf("/api/v1/reports/%d/comments", report.ID), reqBody, userToken)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestUpdateComment(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test users
	user1 := createTestUser(t, db, "commenter3@example.com", "commenter3", "password123", models.RoleCommonUser)
	user1Token, err := generateTestToken(user1, cfg)
	assert.NoError(t, err)

	user2 := createTestUser(t, db, "commenter4@example.com", "commenter4", "password123", models.RoleCommonUser)
	user2Token, err := generateTestToken(user2, cfg)
	assert.NoError(t, err)

	// Create report
	reportRepo := repository.NewReportRepository(db)
	report := &models.Report{
		UserID:      user1.ID,
		Type:        models.ReportTypeRouteIssue,
		Title:       "Report for Update",
		Description: "Description",
		Status:      models.ReportStatusPending,
	}
	reportRepo.Create(report)

	// Create comment
	commentRepo := repository.NewCommentRepository(db)
	comment := &models.Comment{
		ReportID: report.ID,
		UserID:   user1.ID,
		Content:  "Comment to update",
	}
	commentRepo.Create(comment)

	t.Run("update own comment", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"content": "Updated comment content",
		}

		w := makeRequest(router, "PUT", fmt.Sprintf("/api/v1/comments/%d", comment.ID), reqBody, user1Token)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("update another user's comment (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"content": "Unauthorized update",
		}

		w := makeRequest(router, "PUT", fmt.Sprintf("/api/v1/comments/%d", comment.ID), reqBody, user2Token)

		// Should fail - user can't update other user's comments
		assert.True(t, w.Code == http.StatusForbidden || w.Code == http.StatusBadRequest)
	})

	t.Run("update comment without auth (should fail)", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"content": "Unauthorized update",
		}

		w := makeRequest(router, "PUT", fmt.Sprintf("/api/v1/comments/%d", comment.ID), reqBody, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestDeleteComment(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cfg := getTestConfig()
	router := setupTestRouter(db, cfg)

	// Create test users
	user1 := createTestUser(t, db, "commenter5@example.com", "commenter5", "password123", models.RoleCommonUser)
	user1Token, err := generateTestToken(user1, cfg)
	assert.NoError(t, err)

	user2 := createTestUser(t, db, "commenter6@example.com", "commenter6", "password123", models.RoleCommonUser)
	user2Token, err := generateTestToken(user2, cfg)
	assert.NoError(t, err)

	// Create report
	reportRepo := repository.NewReportRepository(db)
	report := &models.Report{
		UserID:      user1.ID,
		Type:        models.ReportTypeRouteIssue,
		Title:       "Report for Delete",
		Description: "Description",
		Status:      models.ReportStatusPending,
	}
	reportRepo.Create(report)

	t.Run("delete own comment", func(t *testing.T) {
		// Create comment to delete
		commentRepo := repository.NewCommentRepository(db)
		comment := &models.Comment{
			ReportID: report.ID,
			UserID:   user1.ID,
			Content:  "Comment to delete",
		}
		commentRepo.Create(comment)

		w := makeRequest(router, "DELETE", fmt.Sprintf("/api/v1/comments/%d", comment.ID), nil, user1Token)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("delete another user's comment (should fail)", func(t *testing.T) {
		// Create comment to delete
		commentRepo := repository.NewCommentRepository(db)
		comment := &models.Comment{
			ReportID: report.ID,
			UserID:   user1.ID,
			Content:  "Comment to delete 2",
		}
		commentRepo.Create(comment)

		w := makeRequest(router, "DELETE", fmt.Sprintf("/api/v1/comments/%d", comment.ID), nil, user2Token)

		// Should fail - user can't delete other user's comments
		assert.True(t, w.Code == http.StatusForbidden || w.Code == http.StatusBadRequest)
	})
}

