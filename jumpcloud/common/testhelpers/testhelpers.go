package testhelpers

import (
	"context"
	"testing"
)

// MockClient represents a mock API client for testing
type MockClient struct {
	APIKey string
	OrgID  string
}

// NewMockClient creates a new mock client for testing
func NewMockClient() *MockClient {
	return &MockClient{
		APIKey: "mock-api-key",
		OrgID:  "mock-org-id",
	}
}

// AccPreCheck validates the necessary test environment variables
func AccPreCheck(t *testing.T) {
	if t != nil {
		t.Log("Running acceptance test pre-checks")
	}
	// In a real implementation, this would check for environment variables
}

// TestCtx returns a context for testing
func TestCtx() context.Context {
	return context.Background()
}
