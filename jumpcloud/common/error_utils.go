package common

import (
	"strings"
)

// IsNotFoundError checks if an error is a "not found" error from the JumpCloud API
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "404") ||
		strings.Contains(err.Error(), "not found") ||
		strings.Contains(err.Error(), "Not Found")
}
