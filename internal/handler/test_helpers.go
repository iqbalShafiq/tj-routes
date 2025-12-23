package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"tj-routes/internal/cache"
	"tj-routes/internal/config"
	"tj-routes/internal/middleware"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
	"tj-routes/internal/service"
	"tj-routes/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var testDB *gorm.DB
var testConfig *config.Config

// getTestConfig returns a test configuration
func getTestConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Port:          "8080",
			Host:          "localhost",
			Environment:   "test",
			AllowedOrigin: "*",
			ReadTimeout:   30,
			WriteTimeout:  30,
		},
		Database: config.DatabaseConfig{
			Host:            getEnv("TEST_DB_HOST", "localhost"),
			Port:            getEnv("TEST_DB_PORT", "5432"),
			User:            getEnv("TEST_DB_USER", "postgres"),
			Password:        getEnv("TEST_DB_PASSWORD", "postgres"),
			Name:            getEnv("TEST_DB_NAME", "tj_routes_test"),
			SSLMode:         "disable",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5,
			RunMigrations:   true,
		},
		JWT: config.JWTConfig{
			Secret:                 "test-secret-key-for-jwt-signing-32chars",
			ExpirationHours:        24,
			RefreshExpirationHours: 168,
		},
		OAuth: config.OAuthConfig{
			GoogleClientID:     "test-client-id",
			GoogleClientSecret: "test-client-secret",
			GoogleRedirectURL:  "http://localhost:8080/api/v1/auth/oauth/google/callback",
		},
		Logging: config.LoggingConfig{
			Level:        "error",
			ErrorLogFile: "",
		},
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			DB:       1, // Use DB 1 for tests to avoid conflicts
			PoolSize: 5,
		},
		Cache: config.CacheConfig{
			Enabled:       false, // Disable cache for tests
			RouteTTL:      30,
			StopTTL:       60,
			VehicleTTL:    15,
			SystemUserTTL: 1440,
		},
		FileStorage: config.FileStorageConfig{
			StorageType:      "local",
			UploadPath:       "./test_uploads",
			MaxPhotoSize:     5 * 1024 * 1024,
			MaxPDFSize:       10 * 1024 * 1024,
			AllowedPhotoMIMEs: []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
			AllowedPDFMIMEs:   []string{"application/pdf"},
		},
		JobQueue: config.JobQueueConfig{
			JobTimeoutMinutes:        5,
			StuckJobThresholdMinutes: 10,
			MaxRetryAttempts:         3,
			Concurrency:              5,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// setupTestDB initializes and migrates the test database
func setupTestDB(t *testing.T) *gorm.DB {
	cfg := getTestConfig()
	db, err := utils.InitDB(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := utils.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Ensure system user exists
	if _, err := utils.EnsureSystemUser(db); err != nil {
		t.Fatalf("Failed to ensure system user: %v", err)
	}

	return db
}

// teardownTestDB cleans up the test database
func teardownTestDB(t *testing.T, db *gorm.DB) {
	if db == nil {
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Logf("Failed to get sql.DB: %v", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		t.Logf("Failed to close database: %v", err)
	}
}

// setupTestRouter creates a Gin router with all handlers and middleware configured for testing
func setupTestRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Initialize cache (noop for tests)
	cacheInstance := cache.NewNoOpCache()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	stopRepo := repository.NewStopRepository(db)
	routeRepo := repository.NewRouteRepository(db)
	routeStopRepo := repository.NewRouteStopRepository(db)
	vehicleRepo := repository.NewVehicleRepository(db)
	reportRepo := repository.NewReportRepository(db)
	routeChangeRepo := repository.NewRouteChangeRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	reactionRepo := repository.NewReactionRepository(db)
	badgeRepo := repository.NewBadgeRepository(db)
	userBadgeRepo := repository.NewUserBadgeRepository(db)

	// Initialize file storage
	baseURL := fmt.Sprintf("http://%s:%s", cfg.Server.Host, cfg.Server.Port)
	fileStorage := utils.NewFileStorage(&cfg.FileStorage, baseURL)

	// Initialize services
	userService := service.NewUserService(userRepo, cfg, cacheInstance)
	stopService := service.NewStopService(stopRepo, cacheInstance, cfg)
	routeService := service.NewRouteService(routeRepo, routeStopRepo, stopRepo, routeChangeRepo, cacheInstance, cfg)
	vehicleService := service.NewVehicleService(vehicleRepo, routeRepo, cacheInstance, cfg)
	reputationService := service.NewReputationService(userRepo)
	badgeService := service.NewBadgeService(badgeRepo, userBadgeRepo, userRepo, reportRepo, commentRepo, reactionRepo)
	reportService := service.NewReportServiceWithReputation(
		reportRepo,
		routeRepo,
		stopRepo,
		reputationService,
		badgeService,
	)
	commentService := service.NewCommentService(commentRepo, reportRepo)
	reactionService := service.NewReactionService(reactionRepo, reportRepo, commentRepo, reputationService)

	// Initialize handlers
	authHandler := NewAuthHandler(userService, cfg)
	stopHandler := NewStopHandler(stopService, fileStorage)
	routeHandler := NewRouteHandler(routeService)
	vehicleHandler := NewVehicleHandler(vehicleService, fileStorage)
	reportHandler := NewReportHandler(reportService, userService, fileStorage, reportRepo)
	userHandler := NewUserHandler(userService)
	docsHandler := NewDocsHandler()
	commentHandler := NewCommentHandler(commentService)
	reactionHandler := NewReactionHandler(reactionService)
	leaderboardHandler := NewLeaderboardHandler(userRepo, badgeService, reputationService)
	bulkUploadHandler := NewBulkUploadHandler(nil) // No bulk upload service for tests

	// Store config in context
	router.Use(func(c *gin.Context) {
		c.Set("config", cfg)
		c.Next()
	})

	// CORS middleware
	router.Use(middleware.CORSMiddleware(cfg))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		if err := utils.CheckDBHealth(db); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "Database connection failed",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	// API Documentation routes
	docs := router.Group("/api/docs")
	{
		docs.GET("", docsHandler.ServeScalarUI)
		docs.GET("/swagger", docsHandler.ServeSwaggerUI)
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
			stops := public.Group("/stops")
			{
				stops.GET("", stopHandler.ListStops)
				stops.GET("/:id", stopHandler.GetStop)
			}

			routes := public.Group("/routes")
			{
				routes.GET("", routeHandler.ListRoutes)
				routes.GET("/:id", routeHandler.GetRoute)
			}

			vehicles := public.Group("/vehicles")
			{
				vehicles.GET("", vehicleHandler.ListVehicles)
				vehicles.GET("/:id", vehicleHandler.GetVehicle)
			}

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
				reports.GET("/:id/comments", commentHandler.GetComments)
				reports.POST("/:id/comments", commentHandler.CreateComment)
				reports.POST("/:id/react", reactionHandler.ReactToReport)
				reports.DELETE("/:id/react", reactionHandler.RemoveReactionFromReport)
			}

			// Comments
			comments := protected.Group("/comments")
			{
				comments.PUT("/:id", commentHandler.UpdateComment)
				comments.DELETE("/:id", commentHandler.DeleteComment)
				comments.POST("/:id/react", reactionHandler.ReactToComment)
				comments.DELETE("/:id/react", reactionHandler.RemoveReactionFromComment)
			}

			// Leaderboard and badges
			leaderboard := protected.Group("/leaderboard")
			{
				leaderboard.GET("", leaderboardHandler.GetLeaderboard)
			}

			badges := protected.Group("/badges")
			{
				badges.GET("", leaderboardHandler.GetAllBadges)
			}

			// Users
			users := protected.Group("/users")
			{
				users.GET("/:id/profile", leaderboardHandler.GetUserProfile)
				adminUsers := users.Group("")
				adminUsers.Use(middleware.RequireAdmin())
				{
					adminUsers.GET("", userHandler.ListUsers)
					adminUsers.GET("/:id", userHandler.GetUser)
					adminUsers.PUT("/:id/role", userHandler.UpdateUserRole)
				}
			}

			// Bulk upload (admin only)
			bulkUpload := protected.Group("/bulk-upload")
			bulkUpload.Use(middleware.RequireAdmin())
			{
				bulkUpload.POST("/:entityType", bulkUploadHandler.UploadCSV)
				bulkUpload.GET("/:id", bulkUploadHandler.GetUploadStatus)
				bulkUpload.GET("", bulkUploadHandler.ListUploads)
			}
		}
	}

	return router
}

// createTestUser creates a test user with the specified role
func createTestUser(t *testing.T, db *gorm.DB, email, username, password string, role models.UserRole) *models.User {
	userRepo := repository.NewUserRepository(db)

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.User{
		Email:    email,
		Username: username,
		Password: &hashedPassword,
		Role:     role,
	}

	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

// generateTestToken generates a JWT token for the given user
func generateTestToken(user *models.User, cfg *config.Config) (string, error) {
	return utils.GenerateToken(user, &cfg.JWT)
}

// makeRequest makes an HTTP request to the router and returns the response
func makeRequest(router *gin.Engine, method, url string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			panic(fmt.Sprintf("Failed to marshal request body: %v", err))
		}
	}

	req := httptest.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

