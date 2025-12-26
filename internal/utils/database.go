package utils

import (
	"time"

	"tj-routes/internal/config"
	"tj-routes/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	// Determine log level based on config
	var logLevel logger.LogLevel
	switch cfg.Logging.Level {
	case "debug":
		logLevel = logger.Info
	case "warn":
		logLevel = logger.Warn
	case "error":
		logLevel = logger.Error
	default:
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Minute)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// CheckDBHealth checks if the database connection is healthy
func CheckDBHealth(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

func AutoMigrate(db *gorm.DB) error {
	// Run standard migrations
	if err := db.AutoMigrate(
		&models.User{},
		&models.Stop{},
		&models.Route{},
		&models.RouteStop{},
		&models.Vehicle{},
		&models.Report{},
		&models.ReportCategory{},
		&models.RouteChange{},
		&models.BulkUploadLog{},
		&models.Comment{},
		&models.Reaction{},
		&models.Badge{},
		&models.UserBadge{},
		&models.UserFollow{},
		&models.Hashtag{},
		&models.ReportHashtag{},
		&models.Forum{},
		&models.ForumPost{},
		&models.ForumMember{},
		&models.ForumPostHashtag{},
	); err != nil {
		return err
	}

	// Enable pg_trgm extension for fuzzy search
	if err := EnablePgTrgmExtension(db); err != nil {
		return err
	}

	// Create trigram indexes for fuzzy search
	if err := CreateTrigramIndexes(db); err != nil {
		return err
	}

	return nil
}

// EnablePgTrgmExtension enables the pg_trgm extension for fuzzy search
func EnablePgTrgmExtension(db *gorm.DB) error {
	return db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error
}

// CreateTrigramIndexes creates GIN indexes on searchable columns for fuzzy search performance
func CreateTrigramIndexes(db *gorm.DB) error {
	indexes := []string{
		// Routes indexes
		"CREATE INDEX IF NOT EXISTS idx_routes_name_trgm ON routes USING GIN (name gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_routes_route_number_trgm ON routes USING GIN (route_number gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_routes_description_trgm ON routes USING GIN (description gin_trgm_ops)",
		// Vehicles indexes
		"CREATE INDEX IF NOT EXISTS idx_vehicles_plate_trgm ON vehicles USING GIN (vehicle_plate gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_vehicles_type_trgm ON vehicles USING GIN (vehicle_type gin_trgm_ops)",
		// Reports indexes
		"CREATE INDEX IF NOT EXISTS idx_reports_title_trgm ON reports USING GIN (title gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_reports_description_trgm ON reports USING GIN (description gin_trgm_ops)",
	}

	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			return err
		}
	}

	return nil
}

// EnsureSystemUser ensures that a system user exists for guest reports
// Returns the system user ID
func EnsureSystemUser(db *gorm.DB) (uint, error) {
	const systemUserEmail = "system@tj-routes.local"
	const systemUsername = "system"

	var systemUser models.User
	err := db.Where("email = ?", systemUserEmail).First(&systemUser).Error
	
	if err == nil {
		// System user already exists
		return systemUser.ID, nil
	}

	if err != gorm.ErrRecordNotFound {
		// Some other error occurred
		return 0, err
	}

	// System user doesn't exist, create it
	systemUser = models.User{
		Email:    systemUserEmail,
		Username: systemUsername,
		Role:     models.RoleCommonUser,
		Password: nil, // No password for system user
	}

	if err := db.Create(&systemUser).Error; err != nil {
		return 0, err
	}

	return systemUser.ID, nil
}