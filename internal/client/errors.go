package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// JumpCloudError represents an error returned by the JumpCloud API
type JumpCloudError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"` // HTTP status code
}

// Error implements the error interface
func (e *JumpCloudError) Error() string {
	return fmt.Sprintf("[%s - %d] %s", e.Code, e.StatusCode, e.Message)
}

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

// ParseJumpCloudError parses an error response from the JumpCloud API
func ParseJumpCloudError(statusCode int, body []byte) *JumpCloudError {
	// Default error in case parsing fails
	defaultError := &JumpCloudError{
		Code:       "UNKNOWN",
		Message:    fmt.Sprintf("Unknown error with status code %d", statusCode),
		StatusCode: statusCode,
	}

	// If there's no body, return default error
	if len(body) == 0 {
		return defaultError
	}

	// Try to parse the error response
	var errorResponse JumpCloudError
	if err := json.Unmarshal(body, &errorResponse); err != nil {
		return defaultError
	}

	// Set the status code
	errorResponse.StatusCode = statusCode

	// If code is empty, set it based on status code
	if errorResponse.Code == "" {
		switch statusCode {
		case http.StatusUnauthorized:
			errorResponse.Code = ERROR_AUTH_FAILED
		case http.StatusForbidden:
			errorResponse.Code = ERROR_PERMISSION_DENIED
		case http.StatusNotFound:
			errorResponse.Code = ERROR_NOT_FOUND
		case http.StatusConflict:
			errorResponse.Code = ERROR_ALREADY_EXISTS
		case http.StatusBadRequest:
			errorResponse.Code = ERROR_INVALID_INPUT
		case http.StatusInternalServerError:
			errorResponse.Code = ERROR_INTERNAL
		case http.StatusServiceUnavailable:
			errorResponse.Code = ERROR_UNAVAILABLE
		case http.StatusGatewayTimeout:
			errorResponse.Code = ERROR_DEADLINE_EXCEEDED
		default:
			errorResponse.Code = "UNKNOWN"
		}
	}

	return &errorResponse
}

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
