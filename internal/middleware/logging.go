package middleware

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"tj-routes/internal/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SetupLogger initializes and returns a zap logger
func SetupLogger(cfg *config.Config) (*zap.Logger, error) {
	var zapConfig zap.Config

	if cfg.Server.Environment == "production" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// Set log level
	switch strings.ToLower(cfg.Logging.Level) {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// Configure error log file if specified
	var cores []zapcore.Core

	// Debug: Check if ErrorLogFile is set (using fmt to avoid logger dependency)
	if cfg.Logging.ErrorLogFile == "" {
		// If empty, check if it should be loaded from env
		// This helps debug config loading issues
	}

	if cfg.Logging.ErrorLogFile != "" {
		// Clean and normalize the path
		logPath := filepath.Clean(cfg.Logging.ErrorLogFile)
		// Convert to absolute path to avoid issues with relative paths
		if !filepath.IsAbs(logPath) {
			absPath, err := filepath.Abs(logPath)
			if err != nil {
				return nil, err
			}
			logPath = absPath
		}

		// Ensure the log file directory exists
		logDir := filepath.Dir(logPath)
		// Create directory if it's not root or current directory
		if logDir != "." && logDir != "" && logDir != filepath.VolumeName(logPath) {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				return nil, err
			}
		}

		// Open the error log file (create if it doesn't exist)
		// Using 0640 permissions to prevent world-readable log files
		errorLogFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
		if err != nil {
			return nil, err
		}
		// Test write to ensure file is writable
		if _, err := errorLogFile.WriteString(""); err != nil {
			errorLogFile.Close()
			return nil, err
		}
		if err := errorLogFile.Sync(); err != nil {
			errorLogFile.Close()
			return nil, err
		}
		// Store the absolute path for later logging (we'll log it after logger is created)
		_ = logPath // Store for potential future use

		// Create a core that writes error-level logs to the file
		errorFileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(zapConfig.EncoderConfig),
			zapcore.AddSync(errorLogFile),
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl >= zapcore.ErrorLevel
			}),
		)
		cores = append(cores, errorFileCore)
	}

	// Create a core that writes all logs to stderr (original behavior)
	stderrCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapConfig.EncoderConfig),
		zapcore.AddSync(os.Stderr),
		zapConfig.Level,
	)
	cores = append(cores, stderrCore)

	// Combine cores: errors go to both file and stderr (if file configured), other logs only to stderr
	combinedCore := zapcore.NewTee(cores...)

	// Create logger with the combined core
	logger := zap.New(combinedCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}

// RequestLoggingMiddleware logs HTTP requests
func RequestLoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store logger in context
		c.Set("logger", logger)

		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		logFields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
			zap.String("user-agent", c.Request.UserAgent()),
		}

		if errorMessage != "" {
			logFields = append(logFields, zap.String("error", errorMessage))
		}

		if statusCode >= 500 {
			logger.Error("HTTP Request", logFields...)
		} else if statusCode >= 400 {
			logger.Warn("HTTP Request", logFields...)
		} else {
			logger.Info("HTTP Request", logFields...)
		}
	}
}
