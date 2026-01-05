package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
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
	// In development, always run migrations unless explicitly disabled
	shouldRunMigrations := cfg.Database.RunMigrations
	if cfg.Server.Environment == "development" && !shouldRunMigrations {
		logger.Info("Development mode: enabling migrations (override DB_RUN_MIGRATIONS=false)")
		shouldRunMigrations = true
	}

	if shouldRunMigrations {
		logger.Info("Running database migrations...")
		if err := utils.AutoMigrate(db); err != nil {
			logger.Fatal("Failed to run migrations", zap.Error(err))
		}
		logger.Info("✓ Database migrations completed successfully")
	} else {
		logger.Warn("⚠ Skipping database migrations (DB_RUN_MIGRATIONS=false)")
		logger.Warn("⚠ If tables are missing, set DB_RUN_MIGRATIONS=true in your .env file")
	}

	// Ensure system user exists for guest reports
	systemUserID, err := utils.EnsureSystemUser(db)
	if err != nil {
		logger.Fatal("Failed to ensure system user exists", zap.Error(err))
	}
	logger.Info("System user ensured", zap.Uint("system_user_id", systemUserID))

	// Initialize Redis cache
	cacheInstance, err := utils.InitRedis(cfg, logger)
	if err != nil {
		logger.Warn("Failed to initialize Redis cache, continuing without cache", zap.Error(err))
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	stopRepo := repository.NewStopRepository(db)
	routeRepo := repository.NewRouteRepository(db)
	routeStopRepo := repository.NewRouteStopRepository(db)
	vehicleRepo := repository.NewVehicleRepository(db)
	reportRepo := repository.NewReportRepository(db)
	reportCategoryRepo := repository.NewReportCategoryRepository(db)
	routeChangeRepo := repository.NewRouteChangeRepository(db)
	bulkUploadLogRepo := repository.NewBulkUploadLogRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	reactionRepo := repository.NewReactionRepository(db)
	badgeRepo := repository.NewBadgeRepository(db)
	userBadgeRepo := repository.NewUserBadgeRepository(db)
	userFollowRepo := repository.NewUserFollowRepository(db)
	hashtagRepo := repository.NewHashtagRepository(db)
	forumRepo := repository.NewForumRepository(db)
	forumPostRepo := repository.NewForumPostRepository(db)
	forumMemberRepo := repository.NewForumMemberRepository(db)
	checkInRepo := repository.NewCheckInRepository(db)
	// User personalized data repositories
	userFavoriteRepo := repository.NewUserFavoriteRepository(db)
	userPlaceRepo := repository.NewUserPlaceRepository(db)
	userRecentViewRepo := repository.NewUserRecentViewRepository(db)
	userSavedNavRepo := repository.NewUserSavedNavigationRepository(db)

	// Initialize file storage (needed for bulk upload service and handlers)
	baseURL := fmt.Sprintf("http://%s:%s", cfg.Server.Host, cfg.Server.Port)
	if cfg.Server.Environment == "production" {
		// In production, use actual domain from config if available
		if allowedOrigin := cfg.Server.AllowedOrigin; allowedOrigin != "*" && allowedOrigin != "" {
			// Extract protocol and domain from allowed origin
			baseURL = allowedOrigin
		}
	}
	fileStorage := utils.NewFileStorage(&cfg.FileStorage, baseURL)
	logger.Info("File storage initialized", zap.String("type", cfg.FileStorage.StorageType), zap.String("path", cfg.FileStorage.UploadPath))

	// Initialize job queue client (always try, even if cache failed)
	// This allows bulk upload to work even if cache is disabled
	var jobQueueClient *utils.JobQueueClient
	jobQueueClient, err = utils.InitJobQueueClient(cfg)
	if err != nil {
		logger.Warn("Failed to initialize job queue client, continuing without background jobs", zap.Error(err))
		jobQueueClient = nil
	} else {
		logger.Info("Job queue client initialized successfully")
	}

	// Initialize services with cache
	userService := service.NewUserService(userRepo, cfg, cacheInstance)
	stopService := service.NewStopService(stopRepo, cacheInstance, cfg)
	routeService := service.NewRouteService(routeRepo, routeStopRepo, stopRepo, routeChangeRepo, reportRepo, forumPostRepo, cacheInstance, cfg)
	vehicleService := service.NewVehicleService(vehicleRepo, routeRepo, cacheInstance, cfg)
	
	// Initialize reputation and badge services
	reputationService := service.NewReputationService(userRepo)
	badgeService := service.NewBadgeServiceWithCheckIn(badgeRepo, userBadgeRepo, userRepo, reportRepo, commentRepo, reactionRepo, checkInRepo)
	
	// Initialize user follow service
	userFollowService := service.NewUserFollowService(userFollowRepo, userRepo)
	
	// Initialize hashtag service
	hashtagService := service.NewHashtagService(hashtagRepo)
	
	// Initialize report category service
	reportCategoryService := service.NewReportCategoryService(reportCategoryRepo)
	
	// Initialize report service with social features (hashtags, follows, reputation)
	reportService := service.NewReportServiceWithSocial(
		reportRepo,
		routeRepo,
		stopRepo,
		hashtagRepo,
		userFollowService,
		reputationService,
		badgeService,
	)
	
	// Initialize forum services
	forumService := service.NewForumService(forumRepo, forumMemberRepo, routeRepo)
	forumPostService := service.NewForumPostService(forumPostRepo, forumRepo, forumMemberRepo, reportRepo)
	
	// Initialize comment and reaction services (with forum post support)
	commentService := service.NewCommentServiceWithForumPost(commentRepo, reportRepo, forumPostRepo)
	reactionService := service.NewReactionServiceWithForumPost(reactionRepo, reportRepo, commentRepo, forumPostRepo, reputationService)

	// Initialize check-in service
	checkInService := service.NewCheckInService(checkInRepo, routeRepo, stopRepo, userRepo, reputationService, badgeService, cacheInstance)

	// Initialize user personalized service
	userPersonalizedService := service.NewUserPersonalizedService(
		userFavoriteRepo,
		userPlaceRepo,
		userRecentViewRepo,
		userSavedNavRepo,
		userRepo,
		routeRepo,
		stopRepo,
	)

	// Initialize bulk upload service
	var bulkUploadService service.BulkUploadService
	if jobQueueClient != nil {
		bulkUploadService = service.NewBulkUploadService(
			bulkUploadLogRepo,
			fileStorage,
			jobQueueClient,
			cfg,
			logger,
		)
	}

	// Initialize job queue server and processor
	var jobQueueServer *utils.JobQueueServer
	var bulkUploadProcessor *service.BulkUploadProcessor
	if jobQueueClient != nil {
		jobQueueServer, err = utils.InitJobQueueServer(cfg, logger)
		if err != nil {
			logger.Warn("Failed to initialize job queue server, continuing without background jobs", zap.Error(err))
			jobQueueServer = nil
		} else {
			logger.Info("Job queue server initialized successfully")
			// Initialize bulk upload processor
			bulkUploadProcessor = service.NewBulkUploadProcessor(
				bulkUploadLogRepo,
				routeRepo,
				stopRepo,
				vehicleRepo,
				cacheInstance,
				cfg,
				logger,
			)

			// Register bulk upload handler
			jobQueueServer.RegisterBulkUploadHandler(bulkUploadProcessor.ProcessBulkUpload)

			// Start job queue server in background
			go func() {
				logger.Info("Starting job queue server")
				if err := jobQueueServer.Start(); err != nil {
					logger.Fatal("Job queue server failed", zap.Error(err))
				}
			}()

			// Recover stuck jobs on startup
			if bulkUploadService != nil {
				if err := bulkUploadService.RecoverStuckJobs(); err != nil {
					logger.Warn("Failed to recover stuck jobs", zap.Error(err))
				} else {
					logger.Info("Job recovery completed")
				}
			}
		}
	}

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userService, cfg, cacheInstance)
	stopHandler := handler.NewStopHandler(stopService, fileStorage)
	routeHandler := handler.NewRouteHandler(routeService)
	vehicleHandler := handler.NewVehicleHandler(vehicleService, fileStorage)
	// Use enhanced report handler with social features
	reportHandler := handler.NewReportHandlerWithSocial(reportService, userService, fileStorage, reportRepo, reactionRepo, userFollowService)
	userHandler := handler.NewUserHandler(userService)
	docsHandler := handler.NewDocsHandler()
	commentHandler := handler.NewCommentHandler(commentService, reactionRepo)
	reactionHandler := handler.NewReactionHandler(reactionService)
	leaderboardHandler := handler.NewLeaderboardHandler(userRepo, badgeService, reputationService)
	userFollowHandler := handler.NewUserFollowHandler(userFollowService)
	hashtagHandler := handler.NewHashtagHandler(hashtagService)
	reportCategoryHandler := handler.NewReportCategoryHandler(reportCategoryService)
	forumHandler := handler.NewForumHandler(forumService)
	forumPostHandler := handler.NewForumPostHandler(forumPostService, fileStorage)
	checkInHandler := handler.NewCheckInHandler(checkInService)

	// Initialize user personalized handler
	userPersonalizedHandler := handler.NewUserPersonalizedHandler(userPersonalizedService)

	// Initialize bulk upload handler (always create, even if service is nil)
	// This ensures routes are registered, but will return error if service unavailable
	var bulkUploadHandler *handler.BulkUploadHandler
	if bulkUploadService != nil {
		bulkUploadHandler = handler.NewBulkUploadHandler(bulkUploadService)
		logger.Info("Bulk upload handler initialized successfully")
	} else {
		logger.Warn("Bulk upload service not available - routes will return service unavailable")
		// Create handler with nil service - it will handle the error gracefully
		bulkUploadHandler = handler.NewBulkUploadHandler(nil)
	}

	// Defensive check - ensure handler is never nil
	if bulkUploadHandler == nil {
		logger.Error("CRITICAL: Bulk upload handler is nil - creating fallback handler")
		bulkUploadHandler = handler.NewBulkUploadHandler(nil)
	}

	// Setup router
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// Request size limit middleware - prevents DoS attacks via large request bodies
	// 10MB limit for most requests
	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10*1024*1024)
		c.Next()
	})

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

	// Serve uploaded files
	router.Static("/uploads", cfg.FileStorage.UploadPath)

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

	// API Documentation routes (public)
	docs := router.Group("/api/docs")
	{
		docs.GET("", docsHandler.ServeScalarUI)
		docs.GET("/swagger", docsHandler.ServeSwaggerUI) // Alternative Swagger UI
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
				routes.GET("/:id/forum", forumHandler.GetForumByRoute)
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
				// Public feed endpoint
				reports.GET("/feed", reportHandler.GetPublicFeed)
				// Trending reports
				reports.GET("/trending", reportHandler.GetTrending)
				// Stories
				reports.GET("/stories", reportHandler.GetStories)
				// Get report details (public access)
				reports.GET("/:id", reportHandler.GetReport)
			}

			// Hashtags (public read access)
			hashtags := public.Group("/hashtags")
			{
				hashtags.GET("/trending", hashtagHandler.GetTrending)
				hashtags.GET("/search", hashtagHandler.SearchHashtags)
				hashtags.GET("/:name/reports", hashtagHandler.GetReportsByHashtag)
			}

			// Report Categories (public read access)
			reportCategories := public.Group("/report-categories")
			{
				reportCategories.GET("", reportCategoryHandler.ListCategories)
			}

			// Forums (public viewing)
			forums := public.Group("/forums")
			{
				forums.GET("/:id", forumHandler.GetForum)
				forums.GET("/:id/posts", forumPostHandler.ListPosts)
				forums.GET("/:id/posts/:postId", forumPostHandler.GetPost)
				forums.GET("/:id/members", forumHandler.GetForumMembers)
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
				// Note: /feed, /trending, /stories, and /:id are handled in public group with OptionalAuthMiddleware
				reports.PUT("/:id/status", middleware.RequireAdmin(), reportHandler.UpdateReportStatus)
				reports.DELETE("/:id", middleware.RequireAdmin(), reportHandler.DeleteReport)
				
				// Report comments
				reports.GET("/:id/comments", commentHandler.GetComments)
				reports.POST("/:id/comments", commentHandler.CreateComment)
				
				// Report reactions
				reports.POST("/:id/react", reactionHandler.ReactToReport)
				reports.DELETE("/:id/react", reactionHandler.RemoveReactionFromReport)
			}
			
			// Comments
			comments := protected.Group("/comments")
			{
				comments.PUT("/:id", commentHandler.UpdateComment)
				comments.DELETE("/:id", commentHandler.DeleteComment)
				
				// Comment reactions
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
				// Public profile endpoint (authenticated users can view any profile)
				users.GET("/:id/profile", leaderboardHandler.GetUserProfile)
				
				// User follow endpoints
				users.POST("/:id/follow", userFollowHandler.FollowUser)
				users.DELETE("/:id/follow", userFollowHandler.UnfollowUser)
				users.GET("/:id/followers", userFollowHandler.GetFollowers)
				users.GET("/:id/following", userFollowHandler.GetFollowing)
				users.GET("/:id/follow-status", userFollowHandler.GetFollowStatus)
				
				// Admin-only endpoints
				adminUsers := users.Group("")
				adminUsers.Use(middleware.RequireAdmin())
				{
					adminUsers.GET("", userHandler.ListUsers)
					adminUsers.GET("/:id", userHandler.GetUser)
					adminUsers.PUT("/:id/role", userHandler.UpdateUserRole)
				}
			}

			// Report Categories (admin create)
			reportCategories := protected.Group("/report-categories")
			reportCategories.Use(middleware.RequireAdmin())
			{
				reportCategories.POST("", reportCategoryHandler.CreateCategory)
			}

			// Bulk upload (admin only)
			// Always register routes, handler will return service unavailable if not initialized
			bulkUpload := protected.Group("/bulk-upload")
			bulkUpload.Use(middleware.RequireAdmin())
			{
				bulkUpload.POST("/:entityType", bulkUploadHandler.UploadCSV)
				bulkUpload.GET("/:id", bulkUploadHandler.GetUploadStatus)
				bulkUpload.GET("", bulkUploadHandler.ListUploads)
			}

			// Forums (protected participation)
			forums := protected.Group("/forums")
			{
				forums.POST("/:id/join", forumHandler.JoinForum)
				forums.DELETE("/:id/leave", forumHandler.LeaveForum)
				forums.GET("/:id/membership", forumHandler.CheckMembership)
				
				forumPosts := forums.Group("/:id/posts")
				{
					forumPosts.POST("", forumPostHandler.CreatePost)
					forumPosts.PUT("/:postId", forumPostHandler.UpdatePost)
					forumPosts.DELETE("/:postId", forumPostHandler.DeletePost)
					forumPosts.POST("/:postId/pin", middleware.RequireAdmin(), forumPostHandler.PinPost)
					forumPosts.DELETE("/:postId/pin", middleware.RequireAdmin(), forumPostHandler.UnpinPost)
				}
			}

			// Forum post comments
			forumPosts := protected.Group("/forum-posts")
			{
				forumPosts.GET("/:id/comments", commentHandler.GetCommentsByForumPost)
				forumPosts.POST("/:id/comments", commentHandler.CreateCommentOnForumPost)

				// Forum post reactions
				forumPosts.POST("/:id/react", reactionHandler.ReactToForumPost)
				forumPosts.DELETE("/:id/react", reactionHandler.RemoveReactionFromForumPost)
			}

			// Journey check-ins (private - user only accesses their own)
			checkIns := protected.Group("/check-ins")
			{
				checkIns.POST("", checkInHandler.CreateCheckIn)
				checkIns.GET("", checkInHandler.ListCheckIns)
				checkIns.GET("/stats", checkInHandler.GetStats)
				checkIns.GET("/:id", checkInHandler.GetCheckIn)
				checkIns.PUT("/:id/complete", checkInHandler.CompleteCheckIn)
				checkIns.PUT("/:id", checkInHandler.UpdateCheckIn)
				checkIns.DELETE("/:id", checkInHandler.DeleteCheckIn)
			}

			// User personalized data (private - user only accesses their own)
			personalized := protected.Group("/users/me/personalized")
			{
				// Favorites
				favorites := personalized.Group("/favorites")
				{
					routes := favorites.Group("/routes")
					{
						routes.GET("", userPersonalizedHandler.GetFavoriteRoutes)
						routes.POST("/:routeId", userPersonalizedHandler.AddFavoriteRoute)
						routes.DELETE("/:routeId", userPersonalizedHandler.RemoveFavoriteRoute)
						routes.GET("/:routeId", userPersonalizedHandler.IsFavoriteRoute)
					}
					stops := favorites.Group("/stops")
					{
						stops.GET("", userPersonalizedHandler.GetFavoriteStops)
						stops.POST("/:stopId", userPersonalizedHandler.AddFavoriteStop)
						stops.DELETE("/:stopId", userPersonalizedHandler.RemoveFavoriteStop)
						stops.GET("/:stopId", userPersonalizedHandler.IsFavoriteStop)
					}
				}

				// Places
				places := personalized.Group("/places")
				{
					places.GET("", userPersonalizedHandler.GetPlaces)
					places.POST("", userPersonalizedHandler.CreatePlace)
					places.GET("/:id", userPersonalizedHandler.GetPlace)
					places.PUT("/:id", userPersonalizedHandler.UpdatePlace)
					places.DELETE("/:id", userPersonalizedHandler.DeletePlace)
				}

				// Recent Views
				recent := personalized.Group("/recent")
				{
					recent.GET("/routes", userPersonalizedHandler.GetRecentRoutes)
					recent.GET("/stops", userPersonalizedHandler.GetRecentStops)
					recent.GET("/navigations", userPersonalizedHandler.GetRecentNavigations)
					recent.POST("/navigations", userPersonalizedHandler.RecordRecentNavigation)
				}

				// Saved Navigations
				savedNavs := personalized.Group("/saved-navigations")
				{
					savedNavs.GET("", userPersonalizedHandler.GetSavedNavigations)
					savedNavs.POST("", userPersonalizedHandler.CreateSavedNavigation)
					savedNavs.PUT("/:id", userPersonalizedHandler.UpdateSavedNavigation)
					savedNavs.DELETE("/:id", userPersonalizedHandler.DeleteSavedNavigation)
				}

				// Analytics
				personalized.GET("/analytics", userPersonalizedHandler.GetAnalytics)
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

		// Check if TLS is enabled
		if cfg.TLS.Enabled {
			// TLS configuration
			tlsConfig := &tls.Config{
				MinVersion:               tls.VersionTLS12,
				CurvePreferences:         []tls.CurveID{tls.CurveP256, tls.X25519},
				PreferServerCipherSuites: true,
			}

			// Set minimum TLS version if specified
			switch cfg.TLS.MinVersion {
			case "1.2":
				tlsConfig.MinVersion = tls.VersionTLS12
			case "1.3":
				tlsConfig.MinVersion = tls.VersionTLS13
			}

			srv.TLSConfig = tlsConfig
			logger.Info("Starting server with TLS",
				zap.String("cert_file", cfg.TLS.CertFile),
				zap.String("key_file", cfg.TLS.KeyFile),
			)
			if err := srv.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile); err != nil && err != http.ErrServerClosed {
				logger.Fatal("Failed to start TLS server", zap.Error(err))
			}
		} else {
			logger.Warn("Starting server without TLS - use a reverse proxy for production")
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal("Failed to start server", zap.Error(err))
			}
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

	// Shutdown job queue server
	if jobQueueServer != nil {
		logger.Info("Shutting down job queue server...")
		jobQueueServer.Shutdown()
	}

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
