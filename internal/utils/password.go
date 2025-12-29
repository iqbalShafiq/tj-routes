package utils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHashCost is the bcrypt cost factor for password hashing
// Using 12 instead of DefaultCost (10) for stronger security
// Each increment roughly doubles the computational cost
const PasswordHashCost = 12

// PasswordValidationError represents password validation errors
type PasswordValidationError struct {
	Messages []string
}

func (e *PasswordValidationError) Error() string {
	return strings.Join(e.Messages, "; ")
}

// ValidatePasswordStrength validates password meets security requirements
// Requirements:
// - At least 8 characters
// - At least 1 uppercase letter
// - At least 1 lowercase letter
// - At least 1 digit
// - At least 1 special character
func ValidatePasswordStrength(password string) error {
	var errors []string

	if len(password) < 8 {
		errors = append(errors, "password must be at least 8 characters long")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?", char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		errors = append(errors, "password must contain at least 1 uppercase letter")
	}
	if !hasLower {
		errors = append(errors, "password must contain at least 1 lowercase letter")
	}
	if !hasDigit {
		errors = append(errors, "password must contain at least 1 digit")
	}
	if !hasSpecial {
		errors = append(errors, "password must contain at least 1 special character (!@#$%^&*()_+-=[]{}|;':\",./<>?)")
	}

	if len(errors) > 0 {
		return &PasswordValidationError{Messages: errors}
	}

	return nil
}

// IsCommonPassword checks if password is in the list of common passwords
func IsCommonPassword(password string) bool {
	commonPasswords := []string{
		"password", "123456", "12345678", "qwerty", "abc123",
		"password123", "admin", "letmein", "welcome",
		"monkey", "dragon", "master", "login", "access",
	}

	lowerPassword := strings.ToLower(password)
	for _, common := range commonPasswords {
		if lowerPassword == common {
			return true
		}
	}

	// Check for obvious patterns
	pattern := regexp.MustCompile(`^(.)\1+$`)
	if pattern.MatchString(password) {
		return true // e.g., "aaaaa", "111111"
	}

	return false
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), PasswordHashCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

