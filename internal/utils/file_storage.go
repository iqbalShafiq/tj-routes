package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"tj-routes/internal/config"
)

// FileStorage interface defines methods for file storage operations
type FileStorage interface {
	SaveFile(file *multipart.FileHeader, entityType, entityID string) (string, error)
	DeleteFile(filePath string) error
	GetFileURL(filePath string) string
	ValidateFile(file *multipart.FileHeader, allowedMIMEs []string, maxSize int64) error
}

// LocalFileStorage implements FileStorage for local filesystem storage
type LocalFileStorage struct {
	config *config.FileStorageConfig
	baseURL string // Base URL for serving files (e.g., "http://localhost:8080/uploads")
}

// NewLocalFileStorage creates a new LocalFileStorage instance
func NewLocalFileStorage(cfg *config.FileStorageConfig, baseURL string) *LocalFileStorage {
	return &LocalFileStorage{
		config:  cfg,
		baseURL: baseURL,
	}
}

// SaveFile saves a file to local storage and returns the file URL
func (s *LocalFileStorage) SaveFile(file *multipart.FileHeader, entityType, entityID string) (string, error) {
	// Validate file first
	allowedMIMEs := s.config.AllowedPhotoMIMEs
	maxSize := s.config.MaxPhotoSize
	
	// Check if it's a PDF
	if file.Header.Get("Content-Type") == "application/pdf" {
		allowedMIMEs = s.config.AllowedPDFMIMEs
		maxSize = s.config.MaxPDFSize
	}
	
	if err := s.ValidateFile(file, allowedMIMEs, maxSize); err != nil {
		return "", err
	}

	// Sanitize filename
	sanitizedFilename := SanitizeFileName(file.Filename)
	
	// Generate unique filename with timestamp
	ext := filepath.Ext(sanitizedFilename)
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%s_%d%s", entityType, timestamp, ext)
	
	// Create directory structure: uploads/{entityType}/{entityID}/
	uploadDir := filepath.Join(s.config.UploadPath, entityType, entityID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Full file path
	filePath := filepath.Join(uploadDir, filename)

	// Open source file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(filePath) // Clean up on error
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Generate relative path for URL (relative to uploads directory)
	relativePath := filepath.Join(entityType, entityID, filename)
	// Use forward slashes for URLs even on Windows
	relativePath = filepath.ToSlash(relativePath)
	
	// Return full URL
	return s.GetFileURL(relativePath), nil
}

// DeleteFile deletes a file from local storage
func (s *LocalFileStorage) DeleteFile(filePath string) error {
	// Extract relative path from URL if it's a full URL
	relativePath := filePath
	if filepath.IsAbs(filePath) || (len(filePath) > 0 && filePath[0] == '/') {
		// Try to extract relative path from URL
		// Assuming URL format: {baseURL}/uploads/{path}
		prefix := s.baseURL + "/uploads/"
		if len(filePath) > len(prefix) && filePath[:len(prefix)] == prefix {
			relativePath = filePath[len(prefix):]
		} else {
			// If it's an absolute path, use it directly
			relativePath = filePath
		}
	}
	
	// Construct full path
	fullPath := filepath.Join(s.config.UploadPath, relativePath)
	
	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}
	
	// Delete file
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	
	return nil
}

// GetFileURL generates a file URL from a relative path
func (s *LocalFileStorage) GetFileURL(relativePath string) string {
	// Ensure relative path uses forward slashes
	relativePath = filepath.ToSlash(relativePath)
	
	// Remove leading slash if present
	if len(relativePath) > 0 && relativePath[0] == '/' {
		relativePath = relativePath[1:]
	}
	
	// Construct URL
	return fmt.Sprintf("%s/uploads/%s", s.baseURL, relativePath)
}

// ValidateFile validates a file using the file validator
func (s *LocalFileStorage) ValidateFile(file *multipart.FileHeader, allowedMIMEs []string, maxSize int64) error {
	return ValidateFile(file, allowedMIMEs, maxSize)
}

// CloudFileStorage implements FileStorage for cloud storage (placeholder for future implementation)
type CloudFileStorage struct {
	config *config.FileStorageConfig
}

// NewCloudFileStorage creates a new CloudFileStorage instance (placeholder)
func NewCloudFileStorage(cfg *config.FileStorageConfig) *CloudFileStorage {
	return &CloudFileStorage{
		config: cfg,
	}
}

// SaveFile saves a file to cloud storage (not implemented yet)
func (s *CloudFileStorage) SaveFile(file *multipart.FileHeader, entityType, entityID string) (string, error) {
	return "", fmt.Errorf("cloud storage not implemented yet")
}

// DeleteFile deletes a file from cloud storage (not implemented yet)
func (s *CloudFileStorage) DeleteFile(filePath string) error {
	return fmt.Errorf("cloud storage not implemented yet")
}

// GetFileURL generates a file URL from cloud storage (not implemented yet)
func (s *CloudFileStorage) GetFileURL(filePath string) string {
	return ""
}

// ValidateFile validates a file using the file validator
func (s *CloudFileStorage) ValidateFile(file *multipart.FileHeader, allowedMIMEs []string, maxSize int64) error {
	return ValidateFile(file, allowedMIMEs, maxSize)
}

// NewFileStorage creates a FileStorage instance based on config
func NewFileStorage(cfg *config.FileStorageConfig, baseURL string) FileStorage {
	if cfg.StorageType == "cloud" {
		return NewCloudFileStorage(cfg)
	}
	return NewLocalFileStorage(cfg, baseURL)
}

