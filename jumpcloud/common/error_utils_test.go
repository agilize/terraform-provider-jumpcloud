package common

import (
	"errors"
	"testing"
)

func TestIsNotFoundError(t *testing.T) {
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
			name:     "404 error",
			err:      errors.New("404 Not Found"),
			expected: true,
		},
		{
			name:     "not found lowercase",
			err:      errors.New("resource not found"),
			expected: true,
		},
		{
			name:     "not found mixed case",
			err:      errors.New("Resource Not Found"),
			expected: true,
		},
		{
			name:     "status code 404",
			err:      errors.New("status: 404"),
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFoundError(tt.err)
			if result != tt.expected {
				t.Errorf("IsNotFoundError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsConflictError(t *testing.T) {
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
			name:     "409 error",
			err:      errors.New("409 Conflict"),
			expected: true,
		},
		{
			name:     "conflict lowercase",
			err:      errors.New("resource conflict"),
			expected: true,
		},
		{
			name:     "conflict mixed case",
			err:      errors.New("Resource Conflict"),
			expected: true,
		},
		{
			name:     "status code 409",
			err:      errors.New("status: 409"),
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsConflictError(tt.err)
			if result != tt.expected {
				t.Errorf("IsConflictError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsBadRequestError(t *testing.T) {
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
			name:     "400 error",
			err:      errors.New("400 Bad Request"),
			expected: true,
		},
		{
			name:     "bad request lowercase",
			err:      errors.New("bad request"),
			expected: true,
		},
		{
			name:     "bad request mixed case",
			err:      errors.New("Bad Request"),
			expected: true,
		},
		{
			name:     "status code 400",
			err:      errors.New("status: 400"),
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBadRequestError(tt.err)
			if result != tt.expected {
				t.Errorf("IsBadRequestError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsUnauthorizedError(t *testing.T) {
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
			name:     "401 error",
			err:      errors.New("401 Unauthorized"),
			expected: true,
		},
		{
			name:     "unauthorized lowercase",
			err:      errors.New("unauthorized"),
			expected: true,
		},
		{
			name:     "unauthorized mixed case",
			err:      errors.New("Unauthorized"),
			expected: true,
		},
		{
			name:     "status code 401",
			err:      errors.New("status: 401"),
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsUnauthorizedError(tt.err)
			if result != tt.expected {
				t.Errorf("IsUnauthorizedError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsForbiddenError(t *testing.T) {
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
			name:     "403 error",
			err:      errors.New("403 Forbidden"),
			expected: true,
		},
		{
			name:     "forbidden lowercase",
			err:      errors.New("forbidden"),
			expected: true,
		},
		{
			name:     "forbidden mixed case",
			err:      errors.New("Forbidden"),
			expected: true,
		},
		{
			name:     "status code 403",
			err:      errors.New("status: 403"),
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsForbiddenError(tt.err)
			if result != tt.expected {
				t.Errorf("IsForbiddenError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseJumpCloudError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedPrefix string
		expectedNil    bool
	}{
		{
			name:        "nil error",
			err:         nil,
			expectedNil: true,
		},
		{
			name:           "not found error",
			err:            errors.New("404 Not Found"),
			expectedPrefix: "404 Not Found",
			expectedNil:    false,
		},
		{
			name:           "other error",
			err:            errors.New("some other error"),
			expectedPrefix: "JumpCloud API error",
			expectedNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseJumpCloudError(tt.err)
			if tt.expectedNil && result != nil {
				t.Errorf("ParseJumpCloudError() = %v, want nil", result)
			} else if !tt.expectedNil && result == nil {
				t.Errorf("ParseJumpCloudError() = nil, want non-nil")
			} else if !tt.expectedNil && !errors.Is(result, tt.err) && result.Error() != tt.expectedPrefix {
				t.Errorf("ParseJumpCloudError() = %v, want error containing %v", result, tt.expectedPrefix)
			}
		})
	}
}
