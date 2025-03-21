package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// JumpCloudError represents an error returned by the JumpCloud API
type JumpCloudError struct {
	// StatusCode is the HTTP status code returned by the JumpCloud API
	StatusCode int

	// Message is the human-readable error message
	Message string

	// Code is the error code, if available
	Code string

	// Raw is the raw error response body
	Raw []byte
}

// Error returns a string representation of the error
func (e *JumpCloudError) Error() string {
	return fmt.Sprintf("JumpCloud API error (status %d): %s", e.StatusCode, e.Message)
}

// ParseJumpCloudError attempts to parse an error response from the JumpCloud API
// It returns a structured JumpCloudError for easier error handling
func ParseJumpCloudError(statusCode int, body []byte) error {
	// Create a base error with the status code and raw body
	jumpCloudErr := &JumpCloudError{
		StatusCode: statusCode,
		Raw:        body,
	}

	// Try to parse the error message from the response body
	var errResponse struct {
		Message string `json:"message"`
		Code    string `json:"code"`
		Error   string `json:"error"`
	}

	if err := json.Unmarshal(body, &errResponse); err == nil {
		// Use the message field if present, otherwise use the error field
		if errResponse.Message != "" {
			jumpCloudErr.Message = errResponse.Message
		} else if errResponse.Error != "" {
			jumpCloudErr.Message = errResponse.Error
		}

		// Set the error code if available
		jumpCloudErr.Code = errResponse.Code
	}

	// If no message was found, create a generic one
	if jumpCloudErr.Message == "" {
		jumpCloudErr.Message = fmt.Sprintf("API request failed with status code %d", statusCode)
	}

	return jumpCloudErr
}

// Common error scenarios
const (
	ErrorNotFound             = "Resource not found"
	ErrorAlreadyExists        = "Resource already exists"
	ErrorInvalidInput         = "Invalid input provided"
	ErrorPermissionDenied     = "Permission denied"
	ErrorAuthenticationFailed = "Authentication failed"
	ErrorInternalServer       = "Internal server error"
	ErrorServiceUnavailable   = "Service unavailable"
)

// Error codes
const (
	// Error codes for authentication issues
	ERROR_AUTH_FAILED       = "AUTH_FAILED"
	ERROR_PERMISSION_DENIED = "PERMISSION_DENIED"

	// Error codes for resource issues
	ERROR_NOT_FOUND      = "NOT_FOUND"
	ERROR_ALREADY_EXISTS = "ALREADY_EXISTS"
	ERROR_INVALID_INPUT  = "INVALID_INPUT"

	// Error codes for system issues
	ERROR_INTERNAL          = "INTERNAL"
	ERROR_UNAVAILABLE       = "UNAVAILABLE"
	ERROR_DEADLINE_EXCEEDED = "DEADLINE_EXCEEDED"
)

// IsAuthError returns true if the error is an authentication error
func IsAuthError(err error) bool {
	if jcErr, ok := err.(*JumpCloudError); ok {
		return jcErr.StatusCode == http.StatusUnauthorized || jcErr.Code == ERROR_AUTH_FAILED
	}
	return false
}

// IsPermissionDenied returns true if the error is a permission denied error
func IsPermissionDenied(err error) bool {
	if jcErr, ok := err.(*JumpCloudError); ok {
		return jcErr.StatusCode == http.StatusForbidden || jcErr.Code == ERROR_PERMISSION_DENIED
	}
	return false
}

// IsNotFound returns true if the error is a not found error
func IsNotFound(err error) bool {
	if jcErr, ok := err.(*JumpCloudError); ok {
		return jcErr.StatusCode == http.StatusNotFound || jcErr.Code == ERROR_NOT_FOUND
	}
	return false
}

// IsAlreadyExists returns true if the error is an already exists error
func IsAlreadyExists(err error) bool {
	if jcErr, ok := err.(*JumpCloudError); ok {
		return jcErr.StatusCode == http.StatusConflict || jcErr.Code == ERROR_ALREADY_EXISTS
	}
	return false
}
