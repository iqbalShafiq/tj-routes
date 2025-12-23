package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	OAuth    OAuthConfig
	Logging  LoggingConfig
	Redis    RedisConfig
	Cache    CacheConfig
}

type ServerConfig struct {
	Port          string
	Host          string
	Environment   string
	AllowedOrigin string // For CORS - comma-separated origins or "*" for all
	ReadTimeout   int    // in seconds
	WriteTimeout  int    // in seconds
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int  // Maximum number of open connections
	MaxIdleConns    int  // Maximum number of idle connections
	ConnMaxLifetime int  // Connection max lifetime in minutes
	RunMigrations   bool // Whether to run AutoMigrate on startup
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

type JWTConfig struct {
	Secret                 string
	ExpirationHours        int
	RefreshExpirationHours int
}

func (j JWTConfig) ExpirationDuration() time.Duration {
	return time.Duration(j.ExpirationHours) * time.Hour
}

func (j JWTConfig) RefreshExpirationDuration() time.Duration {
	return time.Duration(j.RefreshExpirationHours) * time.Hour
}

type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
}

type LoggingConfig struct {
	Level        string
	ErrorLogFile string // Path to error log file (optional, errors will also go to stderr)
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	PoolSize int
}

type CacheConfig struct {
	Enabled           bool
	RouteTTL          int // TTL in minutes
	StopTTL           int // TTL in minutes
	VehicleTTL        int // TTL in minutes
	SystemUserTTL     int // TTL in minutes
}

func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:          getEnv("SERVER_PORT", "8080"),
			Host:          getEnv("SERVER_HOST", "localhost"),
			Environment:   getEnv("ENVIRONMENT", "development"),
			AllowedOrigin: getEnv("ALLOWED_ORIGIN", "*"),
			ReadTimeout:   getEnvAsInt("SERVER_READ_TIMEOUT", 30),
			WriteTimeout:  getEnvAsInt("SERVER_WRITE_TIMEOUT", 30),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			Name:            getEnv("DB_NAME", "tj_routes"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 5),
			RunMigrations:   getEnvAsBool("DB_RUN_MIGRATIONS", true),
		},
		JWT: JWTConfig{
			Secret:                 getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
			ExpirationHours:        getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
			RefreshExpirationHours: getEnvAsInt("JWT_REFRESH_EXPIRATION_HOURS", 168),
		},
		OAuth: OAuthConfig{
			GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/oauth/google/callback"),
		},
		Logging: LoggingConfig{
			Level:        getEnv("LOG_LEVEL", "info"),
			ErrorLogFile: getEnv("ERROR_LOG_FILE", ""),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			PoolSize: getEnvAsInt("REDIS_POOL_SIZE", 10),
		},
		Cache: CacheConfig{
			Enabled:       getEnvAsBool("CACHE_ENABLED", true),
			RouteTTL:      getEnvAsInt("CACHE_ROUTE_TTL", 30),      // 30 minutes
			StopTTL:       getEnvAsInt("CACHE_STOP_TTL", 60),       // 1 hour
			VehicleTTL:    getEnvAsInt("CACHE_VEHICLE_TTL", 15),    // 15 minutes
			SystemUserTTL: getEnvAsInt("CACHE_SYSTEM_USER_TTL", 1440), // 24 hours
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// Validate checks if the configuration is valid for production
func (c *Config) Validate() error {
	if c.Server.Environment == "production" {
		// Validate JWT secret
		if c.JWT.Secret == "" || c.JWT.Secret == "your-super-secret-jwt-key-change-in-production" {
			return fmt.Errorf("JWT_SECRET must be set to a strong secret in production")
		}
		if len(c.JWT.Secret) < 32 {
			return fmt.Errorf("JWT_SECRET must be at least 32 characters long in production")
		}

		// Validate CORS
		if c.Server.AllowedOrigin == "*" {
			return fmt.Errorf("ALLOWED_ORIGIN cannot be '*' in production. Set specific origins")
		}

		// Validate database SSL
		if c.Database.SSLMode == "disable" {
			return fmt.Errorf("DB_SSLMODE must not be 'disable' in production. Use 'require' or 'verify-full'")
		}

		// Validate database password
		if c.Database.Password == "postgres" || c.Database.Password == "" {
			return fmt.Errorf("DB_PASSWORD must be set to a strong password in production")
		}
	}
	return nil
}
