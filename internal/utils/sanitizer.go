package utils

import (
	"regexp"
	"strings"
)

// ContentSanitizer provides XSS protection for user-generated content
type ContentSanitizer struct {
	// Patterns for dangerous content
	scriptPattern     *regexp.Regexp
	eventPattern      *regexp.Regexp
	javascriptPattern *regexp.Regexp
	iframePattern     *regexp.Regexp
	objectPattern     *regexp.Regexp
	embedPattern      *regexp.Regexp
	formPattern       *regexp.Regexp
}

// NewContentSanitizer creates a new content sanitizer
func NewContentSanitizer() *ContentSanitizer {
	return &ContentSanitizer{
		scriptPattern:     regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
		eventPattern:      regexp.MustCompile(`(?i)\s*on\w+\s*=\s*["'][^"']*["']`),
		javascriptPattern: regexp.MustCompile(`(?i)javascript:`),
		iframePattern:     regexp.MustCompile(`(?i)<iframe[^>]*>.*?</iframe>`),
		objectPattern:     regexp.MustCompile(`(?i)<object[^>]*>.*?</object>`),
		embedPattern:      regexp.MustCompile(`(?i)<embed[^>]*>`),
		formPattern:       regexp.MustCompile(`(?i)<form[^>]*>.*?</form>`),
	}
}

// Sanitize removes potentially dangerous content from the input
func (s *ContentSanitizer) Sanitize(content string) string {
	if content == "" {
		return ""
	}

	// Remove script tags and their contents
	content = s.scriptPattern.ReplaceAllString(content, "")

	// Remove event handlers (onclick, onerror, etc.)
	content = s.eventPattern.ReplaceAllString(content, "")

	// Remove javascript: URLs
	content = s.javascriptPattern.ReplaceAllString(content, "")

	// Remove iframe tags
	content = s.iframePattern.ReplaceAllString(content, "")

	// Remove object tags
	content = s.objectPattern.ReplaceAllString(content, "")

	// Remove embed tags
	content = s.embedPattern.ReplaceAllString(content, "")

	// Remove form tags (to prevent phishing)
	content = s.formPattern.ReplaceAllString(content, "")

	// Trim whitespace
	return strings.TrimSpace(content)
}

// SanitizeMultiple sanitizes multiple content fields
func (s *ContentSanitizer) SanitizeMultiple(fields map[string]string) map[string]string {
	result := make(map[string]string)
	for key, value := range fields {
		result[key] = s.Sanitize(value)
	}
	return result
}

// AllowedTags defines which HTML tags are allowed
var AllowedTags = map[string]bool{
	"p":      true,
	"br":     true,
	"b":      true,
	"i":      true,
	"u":      true,
	"em":     true,
	"strong": true,
	"a":      true,
	"ul":     true,
	"ol":     true,
	"li":     true,
	"blockquote": true,
	"code":   true,
	"pre":    true,
	"h1":     true,
	"h2":     true,
	"h3":     true,
	"h4":     true,
	"h5":     true,
	"h6":     true,
}

// SanitizeHTML sanitizes HTML content but preserves allowed tags
// This is more permissive than Sanitize() - use with caution
func (s *ContentSanitizer) SanitizeHTML(content string) string {
	if content == "" {
		return ""
	}

	// First, do basic sanitization to remove dangerous content
	content = s.Sanitize(content)

	// Then escape any remaining HTML
	content = EscapeHTML(content)

	return content
}

// EscapeHTML escapes HTML special characters
func EscapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// IsValidUsername checks if a username is valid
func IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 50 {
		return false
	}

	// Only allow alphanumeric characters, underscores, and hyphens
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return validPattern.MatchString(username)
}

// IsValidEmail checks if an email is valid
func IsValidEmail(email string) bool {
	// Basic email validation
	pattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return pattern.MatchString(email)
}

// TruncateString truncates a string to the specified length
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength]
}

// StripAllHTML removes all HTML tags from content
func StripAllHTML(content string) string {
	if content == "" {
		return ""
	}

	// Remove HTML tags
	stripTags := regexp.MustCompile(`<[^>]*>`)
	content = stripTags.ReplaceAllString(content, "")

	// Decode common HTML entities
	content = strings.ReplaceAll(content, "&nbsp;", " ")
	content = strings.ReplaceAll(content, "&amp;", "&")
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&gt;", ">")
	content = strings.ReplaceAll(content, "&quot;", `"`)
	content = strings.ReplaceAll(content, "&#39;", "'")

	return strings.TrimSpace(content)
}

// ValidateURL validates that a URL is safe
func ValidateURL(url string) bool {
	if url == "" {
		return true // Empty is okay for optional fields
	}

	// Only allow http and https
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return false
	}

	// Block javascript: URLs
	if strings.Contains(strings.ToLower(url), "javascript:") {
		return false
	}

	return true
}
