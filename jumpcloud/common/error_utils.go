package common

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// IsJCError returns true if the error is a JumpCloud API error
func IsJCError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "JumpCloud API error")
}

// IsNotFoundError checks if the error indicates a resource was not found (HTTP 404)
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Check if the error message contains a 404 status code
	return strings.Contains(err.Error(), "404") ||
		strings.Contains(strings.ToLower(err.Error()), "not found")
}

// IsNotFoundStatus checks if the HTTP status code is 404 Not Found
func IsNotFoundStatus(statusCode int) bool {
	return statusCode == http.StatusNotFound
}

// ParseJumpCloudError parses the error from the JumpCloud API
func ParseJumpCloudError(err error) error {
	if err == nil {
		return nil
	}

	// If it's a not found error, return a standardized error
	if IsNotFoundError(err) {
		return errors.New("404 Not Found")
	}

	// Return the original error with a prefix
	return fmt.Errorf("JumpCloud API error: %w", err)
}

// IsConflictError checks if the error is a 409 Conflict error
func IsConflictError(err error) bool {
	if err == nil {
		return false
	}

	// Check for conflict in error message
	if strings.Contains(strings.ToLower(err.Error()), "conflict") {
		return true
	}

	// Check for HTTP status code in error message
	if strings.Contains(err.Error(), "409") {
		return true
	}

	return false
}

// IsBadRequestError checks if the error is a 400 Bad Request error
func IsBadRequestError(err error) bool {
	if err == nil {
		return false
	}

	// Check for bad request in error message
	if strings.Contains(strings.ToLower(err.Error()), "bad request") {
		return true
	}

	// Check for HTTP status code in error message
	if strings.Contains(err.Error(), "400") {
		return true
	}

	return false
}

// IsUnauthorizedError checks if the error is a 401 Unauthorized error
func IsUnauthorizedError(err error) bool {
	if err == nil {
		return false
	}

	// Check for unauthorized in error message
	if strings.Contains(strings.ToLower(err.Error()), "unauthorized") {
		return true
	}

	// Check for HTTP status code in error message
	if strings.Contains(err.Error(), "401") {
		return true
	}

	return false
}

// IsForbiddenError checks if the error is a 403 Forbidden error
func IsForbiddenError(err error) bool {
	if err == nil {
		return false
	}

	// Check for forbidden in error message
	if strings.Contains(strings.ToLower(err.Error()), "forbidden") {
		return true
	}

	// Check for HTTP status code in error message
	if strings.Contains(err.Error(), "403") {
		return true
	}

	return false
}
