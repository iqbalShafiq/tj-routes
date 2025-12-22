package handler

import (
	"net/http"
	"strings"

	"tj-routes/internal/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

func SuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Data:    data,
	})
}

// sanitizeError sanitizes error messages for production
func sanitizeError(c *gin.Context, err error, cfg *config.Config) string {
	if err == nil {
		return "An error occurred"
	}

	errMsg := err.Error()

	// In production, don't expose internal error details
	if cfg.Server.Environment == "production" {
		// Log the full error (will be handled by logging middleware)
		if logger, ok := c.Get("logger"); ok {
			if zapLogger, ok := logger.(*zap.Logger); ok {
				zapLogger.Error("Internal error", zap.Error(err))
			}
		}

		// Return generic messages based on error type
		errMsgLower := strings.ToLower(errMsg)
		if strings.Contains(errMsgLower, "database") || strings.Contains(errMsgLower, "sql") {
			return "Database error occurred"
		}
		if strings.Contains(errMsgLower, "connection") {
			return "Connection error occurred"
		}
		if strings.Contains(errMsgLower, "not found") {
			return "Resource not found"
		}
		if strings.Contains(errMsgLower, "unauthorized") || strings.Contains(errMsgLower, "permission") {
			return "Unauthorized access"
		}
		if strings.Contains(errMsgLower, "validation") || strings.Contains(errMsgLower, "invalid") {
			return "Invalid request data"
		}

		// Generic error for production
		return "An error occurred. Please try again later."
	}

	// In development, return full error
	return errMsg
}

func ErrorResponse(c *gin.Context, statusCode int, err error) {
	var cfg *config.Config
	if configVal, exists := c.Get("config"); exists {
		cfg = configVal.(*config.Config)
	} else {
		// Fallback if config not in context
		cfg = &config.Config{
			Server: config.ServerConfig{Environment: "development"},
		}
	}

	errorMsg := sanitizeError(c, err, cfg)
	c.JSON(statusCode, Response{
		Success: false,
		Error:   errorMsg,
	})
}

func MessageResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
	})
}

func BadRequest(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusBadRequest, err)
}

func Unauthorized(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusUnauthorized, err)
}

func Forbidden(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusForbidden, err)
}

func NotFound(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusNotFound, err)
}

func InternalServerError(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusInternalServerError, err)
}

