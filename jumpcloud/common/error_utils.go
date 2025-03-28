package common

import (
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
		strings.Contains(err.Error(), "not found") ||
		strings.Contains(err.Error(), "Not Found")
}

// IsNotFoundStatus checks if the HTTP status code is 404 Not Found
func IsNotFoundStatus(statusCode int) bool {
	return statusCode == http.StatusNotFound
}
