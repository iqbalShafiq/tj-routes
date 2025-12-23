package main

import (
	"bufio"
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"tj-routes/internal/config"
	"tj-routes/internal/handler"
	"tj-routes/internal/middleware"
	"tj-routes/internal/repository"
	"tj-routes/internal/service"
	"tj-routes/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// loadEnvFile reads .env file and sets environment variables
func loadEnvFile() {
	wd, _ := os.Getwd()
	envPaths := []string{
		filepath.Join(wd, ".env"), // Absolute path
		".env",                    // Relative path
	}

	var envFile *os.File
	var err error
	for _, envPath := range envPaths {
		envFile, err = os.Open(envPath)
		if err == nil {
			break
		}
	}

	if envFile == nil {
		return // .env file not found, skip loading
	}
	defer envFile.Close()

	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			value = strings.Trim(value, `"'`)
			if key != "" && value != "" {
				// Only set if not already set in environment
				if os.Getenv(key) == "" {
					os.Setenv(key, value)
				}
			}
		}
	}
}

func main() {
	// Load environment variables from .env file
	loadEnvFile()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		zap.L().Fatal("Failed to load config", zap.Error(err))
	}

	// Validate configuration (especially for production)
	if err := cfg.Validate(); err != nil {
		zap.L().Fatal("Configuration validation failed", zap.Error(err))
	}

	// Setup structured logger
	logger, err := middleware.SetupLogger(cfg)
	if err != nil {
		zap.L().Fatal("Failed to setup logger", zap.Error(err))
	}
	defer logger.Sync()

	// Log the error log file configuration for debugging (concise version)
	envValue := os.Getenv("ERROR_LOG_FILE")
	currentWd, _ := os.Getwd()

	// Check if other env vars are loaded to verify godotenv is working
	testEnvVar := os.Getenv("SERVER_PORT")

	logger.Info("Environment check",
		zap.String("working_dir", currentWd),
		zap.String("env_file_exists", func() string {
			if _, err := os.Stat(".env"); err == nil {
				return "yes"
			}
			return "no"
		}()),
		zap.String("ERROR_LOG_FILE_env", envValue),
		zap.String("ERROR_LOG_FILE_config", cfg.Logging.ErrorLogFile),
		zap.String("SERVER_PORT_env", testEnvVar),
		zap.String("godotenv_working", func() string {
			if testEnvVar != "" && testEnvVar != "8080" {
				return "yes (custom value loaded)"
			} else if testEnvVar == "8080" {
				return "maybe (could be default)"
			}
			return "unknown"
		}()),
	)
	if cfg.Logging.ErrorLogFile != "" {
		absPath, _ := filepath.Abs(cfg.Logging.ErrorLogFile)
		logger.Info("✓ Error log file configured",
			zap.String("path", absPath))
	} else {
		logger.Warn("✗ ERROR_LOG_FILE not set - file logging disabled",
			zap.String("env", envValue),
			zap.String("config", cfg.Logging.ErrorLogFile),
			zap.String("hint", "Check if .env file is in: "+currentWd))
	}

	// Replace global logger
	zap.ReplaceGlobals(logger)
	logger.Info("Starting application", zap.String("environment", cfg.Server.Environment))

	// Initialize database
	db, err := utils.InitDB(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	logger.Info("Database connected successfully")

	// Run migrations conditionally
	if cfg.Database.RunMigrations {
		if err := utils.AutoMigrate(db); err != nil {
			logger.Fatal("Failed to run migrations", zap.Error(err))
		}
		logger.Info("Database migrations completed")
	} else {
		logger.Info("Skipping database migrations (DB_RUN_MIGRATIONS=false)")
	}

	// Ensure system user exists for guest reports
	systemUserID, err := utils.EnsureSystemUser(db)
	if err != nil {
		logger.Fatal("Failed to ensure system user exists", zap.Error(err))
	}
	logger.Info("System user ensured", zap.Uint("system_user_id", systemUserID))

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	stopRepo := repository.NewStopRepository(db)
	routeRepo := repository.NewRouteRepository(db)
	routeStopRepo := repository.NewRouteStopRepository(db)
	vehicleRepo := repository.NewVehicleRepository(db)
	reportRepo := repository.NewReportRepository(db)
	routeChangeRepo := repository.NewRouteChangeRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo, cfg)
	stopService := service.NewStopService(stopRepo)
	routeService := service.NewRouteService(routeRepo, routeStopRepo, stopRepo, routeChangeRepo)
	vehicleService := service.NewVehicleService(vehicleRepo, routeRepo)
	reportService := service.NewReportService(reportRepo, routeRepo, stopRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userService, cfg)
	stopHandler := handler.NewStopHandler(stopService)
	routeHandler := handler.NewRouteHandler(routeService)
	vehicleHandler := handler.NewVehicleHandler(vehicleService)
	reportHandler := handler.NewReportHandler(reportService, userService)
	userHandler := handler.NewUserHandler(userService)
	docsHandler := handler.NewDocsHandler()

	// Setup router
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// Recovery middleware (with custom recovery)
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Error("Panic recovered",
			zap.Any("error", recovered),
			zap.String("path", c.Request.URL.Path),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Internal server error",
		})
		c.Abort()
	}))

	// Request logging middleware
	router.Use(middleware.RequestLoggingMiddleware(logger))

	// Store config and logger in context for handlers
	router.Use(func(c *gin.Context) {
		c.Set("config", cfg)
		c.Set("logger", logger)
		c.Next()
	})

	// CORS middleware (secure)
	router.Use(middleware.CORSMiddleware(cfg))

	// Rate limiting middleware
	router.Use(middleware.RateLimitMiddleware(100, 200)) // 100 req/s, burst 200

	// Health check with database
	router.GET("/health", func(c *gin.Context) {
		// Check database connectivity
		if err := utils.CheckDBHealth(db); err != nil {
			logger.Warn("Health check failed - database unreachable", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "Database connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	// Test endpoint for error logging (temporary - remove in production)
	router.GET("/test-error", func(c *gin.Context) {
		logger.Error("Test error log",
			zap.String("message", "This is a test error to verify file logging"),
			zap.String("endpoint", "/test-error"),
			zap.String("timestamp", time.Now().Format(time.RFC3339)),
		)
		// Force sync to ensure log is written immediately
		logger.Sync()

		// Check if log file exists
		logFileExists := false
		logFilePath := ""
		if cfg.Logging.ErrorLogFile != "" {
			absPath, _ := filepath.Abs(cfg.Logging.ErrorLogFile)
			logFilePath = absPath
			if _, err := os.Stat(absPath); err == nil {
				logFileExists = true
			}
		}

		envValue := os.Getenv("ERROR_LOG_FILE")
		wd, _ := os.Getwd()

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Test error - check logs/errors.log file",
			"message": "This endpoint is for testing error file logging only",
			"debug": gin.H{
				"error_log_file_from_env":    envValue,
				"error_log_file_from_config": cfg.Logging.ErrorLogFile,
				"working_directory":          wd,
				"log_file_path":              logFilePath,
				"log_file_exists":            logFileExists,
				"log_file_directory_exists": func() bool {
					if logFilePath != "" {
						dir := filepath.Dir(logFilePath)
						_, err := os.Stat(dir)
						return err == nil
					}
					return false
				}(),
			},
		})
	})

	// API Documentation routes (public)
	docs := router.Group("/api/docs")
	{
		docs.GET("", docsHandler.ServeScalarUI)
		docs.GET("/openapi.yaml", docsHandler.ServeOpenAPISpec)
	}

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/oauth/:provider", authHandler.OAuthInitiate)
			auth.GET("/oauth/:provider/callback", authHandler.OAuthCallback)
		}

		// Public routes (guest-accessible with optional auth)
		public := v1.Group("")
		public.Use(middleware.OptionalAuthMiddleware(cfg))
		{
			// Stops (read for guests, write requires auth+admin)
			stops := public.Group("/stops")
			{
				stops.GET("", stopHandler.ListStops)
				stops.GET("/:id", stopHandler.GetStop)
			}

			// Routes (read for guests, write requires auth+admin)
			routes := public.Group("/routes")
			{
				routes.GET("", routeHandler.ListRoutes)
				routes.GET("/:id", routeHandler.GetRoute)
			}

			// Vehicles (read for guests, write requires auth+admin)
			vehicles := public.Group("/vehicles")
			{
				vehicles.GET("", vehicleHandler.ListVehicles)
				vehicles.GET("/:id", vehicleHandler.GetVehicle)
			}

			// Reports (create for guests and authenticated users)
			reports := public.Group("/reports")
			{
				reports.POST("", reportHandler.CreateReport)
			}
		}

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			// Stops (write operations require admin)
			stops := protected.Group("/stops")
			{
				stops.POST("", middleware.RequireAdmin(), stopHandler.CreateStop)
				stops.PUT("/:id", middleware.RequireAdmin(), stopHandler.UpdateStop)
				stops.DELETE("/:id", middleware.RequireAdmin(), stopHandler.DeleteStop)
			}

			// Routes (write operations require admin)
			routes := protected.Group("/routes")
			{
				routes.POST("", middleware.RequireAdmin(), routeHandler.CreateRoute)
				routes.PUT("/:id", middleware.RequireAdmin(), routeHandler.UpdateRoute)
				routes.PUT("/:id/stops", middleware.RequireAdmin(), routeHandler.UpdateRouteStops)
				routes.DELETE("/:id", middleware.RequireAdmin(), routeHandler.DeleteRoute)
			}

			// Vehicles (write operations require admin)
			vehicles := protected.Group("/vehicles")
			{
				vehicles.POST("", middleware.RequireAdmin(), vehicleHandler.CreateVehicle)
				vehicles.PUT("/:id", middleware.RequireAdmin(), vehicleHandler.UpdateVehicle)
				vehicles.DELETE("/:id", middleware.RequireAdmin(), vehicleHandler.DeleteVehicle)
			}

			// Reports (read/update/delete for authenticated users and admins)
			reports := protected.Group("/reports")
			{
				reports.GET("", reportHandler.ListReports)
				reports.GET("/:id", reportHandler.GetReport)
				reports.PUT("/:id/status", middleware.RequireAdmin(), reportHandler.UpdateReportStatus)
				reports.DELETE("/:id", middleware.RequireAdmin(), reportHandler.DeleteReport)
			}

			// Users (admin only)
			users := protected.Group("/users")
			users.Use(middleware.RequireAdmin())
			{
				users.GET("", userHandler.ListUsers)
				users.GET("/:id", userHandler.GetUser)
				users.PUT("/:id/role", userHandler.UpdateUserRole)
			}
		}
	}

	// Create HTTP server with timeouts
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server starting", zap.String("address", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
