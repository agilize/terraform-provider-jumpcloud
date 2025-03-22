package unit

import (
	"testing"

	"registry.terraform.io/agilize/jumpcloud/pkg/apiclient"
)

func TestNewClient(t *testing.T) {
	// Setup
	config := &apiclient.Config{
		APIKey: "test-api-key",
		OrgID:  "test-org-id",
		APIURL: "https://test.jumpcloud.com/api",
	}

	// Execute
	client := apiclient.NewClient(config)

	// Verify
	if client.GetApiKey() != "test-api-key" {
		t.Errorf("Expected API key to be 'test-api-key', got '%s'", client.GetApiKey())
	}

	if client.GetOrgID() != "test-org-id" {
		t.Errorf("Expected Org ID to be 'test-org-id', got '%s'", client.GetOrgID())
	}

	if client.APIURL != "https://test.jumpcloud.com/api" {
		t.Errorf("Expected API URL to be 'https://test.jumpcloud.com/api', got '%s'", client.APIURL)
	}
}

func TestNewClientWithDefaults(t *testing.T) {
	// Setup
	config := &apiclient.Config{
		APIKey: "test-api-key",
		// No OrgID, APIURL, or Version specified - should use defaults
	}

	// Execute
	client := apiclient.NewClient(config)

	// Verify
	if client.GetApiKey() != "test-api-key" {
		t.Errorf("Expected API key to be 'test-api-key', got '%s'", client.GetApiKey())
	}

	if client.GetOrgID() != "" {
		t.Errorf("Expected Org ID to be empty, got '%s'", client.GetOrgID())
	}

	if client.APIURL != "https://console.jumpcloud.com/api" {
		t.Errorf("Expected API URL to be 'https://console.jumpcloud.com/api', got '%s'", client.APIURL)
	}

	if client.Version != apiclient.V2 {
		t.Errorf("Expected Version to be V2, got '%s'", client.Version)
	}
}
