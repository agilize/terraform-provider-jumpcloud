package provider

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"registry.terraform.io/agilize/jumpcloud/internal/client"
)

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "JumpCloud 404 error",
			err:      &client.JumpCloudError{StatusCode: http.StatusNotFound},
			expected: true,
		},
		{
			name:     "string with 404",
			err:      fmt.Errorf("received 404 status code"),
			expected: true,
		},
		{
			name:     "string with not found",
			err:      fmt.Errorf("resource not found"),
			expected: true,
		},
		{
			name:     "unrelated error",
			err:      fmt.Errorf("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFound(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsUnauthorized(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "JumpCloud 401 error",
			err:      &client.JumpCloudError{StatusCode: http.StatusUnauthorized},
			expected: true,
		},
		{
			name:     "string with unauthorized",
			err:      fmt.Errorf("request unauthorized"),
			expected: true,
		},
		{
			name:     "string with invalid token",
			err:      fmt.Errorf("invalid token provided"),
			expected: true,
		},
		{
			name:     "unrelated error",
			err:      fmt.Errorf("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsUnauthorized(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsForbidden(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "JumpCloud 403 error",
			err:      &client.JumpCloudError{StatusCode: http.StatusForbidden},
			expected: true,
		},
		{
			name:     "string with forbidden",
			err:      fmt.Errorf("access forbidden"),
			expected: true,
		},
		{
			name:     "string with access denied",
			err:      fmt.Errorf("access denied to resource"),
			expected: true,
		},
		{
			name:     "unrelated error",
			err:      fmt.Errorf("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsForbidden(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsConflict(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "JumpCloud 409 error",
			err:      &client.JumpCloudError{StatusCode: http.StatusConflict},
			expected: true,
		},
		{
			name:     "string with conflict",
			err:      fmt.Errorf("resource conflict"),
			expected: true,
		},
		{
			name:     "string with already exists",
			err:      fmt.Errorf("resource already exists"),
			expected: true,
		},
		{
			name:     "unrelated error",
			err:      fmt.Errorf("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsConflict(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsBadRequest(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "JumpCloud 400 error",
			err:      &client.JumpCloudError{StatusCode: http.StatusBadRequest},
			expected: true,
		},
		{
			name:     "string with bad request",
			err:      fmt.Errorf("bad request received"),
			expected: true,
		},
		{
			name:     "string with invalid request",
			err:      fmt.Errorf("invalid request format"),
			expected: true,
		},
		{
			name:     "unrelated error",
			err:      fmt.Errorf("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBadRequest(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "JumpCloud 500 error",
			err:      &client.JumpCloudError{StatusCode: http.StatusInternalServerError},
			expected: true,
		},
		{
			name:     "JumpCloud 502 error",
			err:      &client.JumpCloudError{StatusCode: http.StatusBadGateway},
			expected: true,
		},
		{
			name:     "JumpCloud 503 error",
			err:      &client.JumpCloudError{StatusCode: http.StatusServiceUnavailable},
			expected: true,
		},
		{
			name:     "string with timeout",
			err:      fmt.Errorf("request timeout"),
			expected: true,
		},
		{
			name:     "string with temporary",
			err:      fmt.Errorf("temporary failure"),
			expected: true,
		},
		{
			name:     "string with retry",
			err:      fmt.Errorf("please retry later"),
			expected: true,
		},
		{
			name:     "unrelated error",
			err:      fmt.Errorf("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryable(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
