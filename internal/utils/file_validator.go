package utils

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidMIMEType    = errors.New("invalid MIME type")
	ErrFileTooLarge        = errors.New("file size exceeds maximum allowed")
	ErrInvalidFileExtension = errors.New("invalid file extension")
	ErrInvalidFileContent   = errors.New("invalid file content (magic number mismatch)")
)

// ValidateFile validates a file based on MIME type, size, extension, and magic numbers
func ValidateFile(file *multipart.FileHeader, allowedMIMEs []string, maxSize int64) error {
	// Validate file size
	if file.Size > maxSize {
		return fmt.Errorf("%w: file size %d exceeds maximum %d bytes", ErrFileTooLarge, file.Size, maxSize)
	}

	// Validate MIME type
	mimeType := file.Header.Get("Content-Type")
	if !isAllowedMIME(mimeType, allowedMIMEs) {
		return fmt.Errorf("%w: %s not in allowed list: %v", ErrInvalidMIMEType, mimeType, allowedMIMEs)
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isAllowedExtension(ext, mimeType) {
		return fmt.Errorf("%w: %s not allowed for MIME type %s", ErrInvalidFileExtension, ext, mimeType)
	}

	// Validate magic number (file content)
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Read first few bytes to check magic number
	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && n == 0 {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if !validateMagicNumber(buffer[:n], mimeType) {
		return fmt.Errorf("%w: file content does not match MIME type %s", ErrInvalidFileContent, mimeType)
	}

	return nil
}

// isAllowedMIME checks if the MIME type is in the allowed list
func isAllowedMIME(mimeType string, allowedMIMEs []string) bool {
	for _, allowed := range allowedMIMEs {
		if mimeType == allowed {
			return true
		}
	}
	return false
}

// isAllowedExtension checks if the file extension matches the MIME type
func isAllowedExtension(ext string, mimeType string) bool {
	extMap := map[string][]string{
		"image/jpeg":      {".jpg", ".jpeg"},
		"image/png":       {".png"},
		"image/webp":      {".webp"},
		"image/gif":       {".gif"},
		"application/pdf": {".pdf"},
	}

	allowedExts, ok := extMap[mimeType]
	if !ok {
		return false
	}

	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			return true
		}
	}
	return false
}

// validateMagicNumber checks if the file content matches the expected magic number
func validateMagicNumber(content []byte, mimeType string) bool {
	if len(content) < 4 {
		return false
	}

	magicNumbers := map[string][]byte{
		"image/jpeg":      {0xFF, 0xD8, 0xFF}, // JPEG: FF D8 FF
		"image/png":       {0x89, 0x50, 0x4E, 0x47}, // PNG: 89 50 4E 47
		"image/webp":      {0x52, 0x49, 0x46, 0x46}, // WebP: RIFF (starts with RIFF)
		"image/gif":       {0x47, 0x49, 0x46, 0x38}, // GIF: GIF8
		"application/pdf": {0x25, 0x50, 0x44, 0x46}, // PDF: %PDF
	}

	expectedMagic, ok := magicNumbers[mimeType]
	if !ok {
		return false
	}

	// Special handling for WebP (RIFF...WEBP)
	if mimeType == "image/webp" {
		if len(content) < 12 {
			return false
		}
		// Check for RIFF header
		if !bytes.Equal(content[0:4], []byte{0x52, 0x49, 0x46, 0x46}) {
			return false
		}
		// Check for WEBP in bytes 8-11
		return bytes.Equal(content[8:12], []byte("WEBP"))
	}

	// For other types, check the magic number
	return bytes.HasPrefix(content, expectedMagic)
}

// SanitizeFileName sanitizes a filename to prevent path traversal attacks
func SanitizeFileName(filename string) string {
	// Remove path separators and dangerous characters
	filename = filepath.Base(filename) // Remove any directory components
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	filename = strings.ReplaceAll(filename, "\x00", "") // Remove null bytes
	
	// Limit filename length
	if len(filename) > 255 {
		ext := filepath.Ext(filename)
		name := filename[:255-len(ext)]
		filename = name + ext
	}
	
	return filename
}

