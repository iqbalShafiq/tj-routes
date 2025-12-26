package cache

import (
	"fmt"
	"strconv"
)

// Cache key generation utilities

// RouteKey generates a cache key for a single route
func RouteKey(id uint) string {
	return fmt.Sprintf("route:%d", id)
}

// RouteStatsKey generates a cache key for route statistics
func RouteStatsKey(id uint) string {
	return fmt.Sprintf("route:stats:%d", id)
}

// RouteListKey generates a cache key for route list with filters
func RouteListKey(page, limit int, status, routeNumber, search string) string {
	return fmt.Sprintf("route:list:%d:%d:%s:%s:%s", page, limit, status, routeNumber, search)
}

// StopKey generates a cache key for a single stop
func StopKey(id uint) string {
	return fmt.Sprintf("stop:%d", id)
}

// StopListKey generates a cache key for stop list with filters
func StopListKey(page, limit int, status, stopType, city string) string {
	return fmt.Sprintf("stop:list:%d:%d:%s:%s:%s", page, limit, status, stopType, city)
}

// VehicleKey generates a cache key for a single vehicle
func VehicleKey(id uint) string {
	return fmt.Sprintf("vehicle:%d", id)
}

// VehicleListKey generates a cache key for vehicle list with filters
func VehicleListKey(page, limit int, status string, routeID uint, search string) string {
	return fmt.Sprintf("vehicle:list:%d:%d:%s:%d:%s", page, limit, status, routeID, search)
}

// SystemUserKey generates a cache key for system user
func SystemUserKey() string {
	return "system:user"
}

// RoutePattern returns the pattern for all route keys
func RoutePattern() string {
	return "route:*"
}

// StopPattern returns the pattern for all stop keys
func StopPattern() string {
	return "stop:*"
}

// VehiclePattern returns the pattern for all vehicle keys
func VehiclePattern() string {
	return "vehicle:*"
}

// Helper function to convert uint to string safely
func uintToString(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}

