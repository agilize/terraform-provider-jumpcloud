package integration

import (
	"os"
	"testing"

	"registry.terraform.io/agilize/jumpcloud/pkg/apiclient"
)

// TestClientIntegration performs integration tests with the actual JumpCloud API
// To run these tests, set the JUMPCLOUD_API_KEY and JUMPCLOUD_ORG_ID environment variables
func TestClientIntegration(t *testing.T) {
	// Skip integration tests if not specifically enabled
	if os.Getenv("JUMPCLOUD_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests. Set JUMPCLOUD_INTEGRATION_TESTS=true to enable.")
	}

	// Get API key from environment
	apiKey := os.Getenv("JUMPCLOUD_API_KEY")
	if apiKey == "" {
		t.Fatal("JUMPCLOUD_API_KEY environment variable must be set to run integration tests")
	}

	// Get organization ID from environment (optional)
	orgID := os.Getenv("JUMPCLOUD_ORG_ID")

	// Create client with real credentials
	client := apiclient.NewClient(&apiclient.Config{
		APIKey: apiKey,
		OrgID:  orgID,
	})

	// Test case: Get system users
	t.Run("GetSystemUsers", func(t *testing.T) {
		resp, err := client.GetV2("/systemusers")
		if err != nil {
			t.Fatalf("Failed to get system users: %v", err)
		}

		if len(resp) == 0 {
			t.Errorf("Expected non-empty response for system users")
		}
	})

	// Test case: Get systems
	t.Run("GetSystems", func(t *testing.T) {
		resp, err := client.GetV2("/systems")
		if err != nil {
			t.Fatalf("Failed to get systems: %v", err)
		}

		if len(resp) == 0 {
			t.Errorf("Expected non-empty response for systems")
		}
	})

	// Test case: Error handling for non-existent resource
	t.Run("ErrorHandling", func(t *testing.T) {
		_, err := client.GetV2("/nonexistentresource")
		if err == nil {
			t.Errorf("Expected error for non-existent resource, got nil")
		}

		if !apiclient.IsNotFound(err) {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})
}
