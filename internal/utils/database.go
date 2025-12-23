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
	return db.AutoMigrate(
		&models.User{},
		&models.Stop{},
		&models.Route{},
		&models.RouteStop{},
		&models.Vehicle{},
		&models.Report{},
		&models.RouteChange{},
		&models.BulkUploadLog{},
	)
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