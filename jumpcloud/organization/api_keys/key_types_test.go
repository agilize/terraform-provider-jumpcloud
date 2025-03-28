package api_keys

import (
	"testing"
	"time"
)

func TestAPIKey(t *testing.T) {
	// Test creating an APIKey instance
	expiryTime := time.Now().Add(24 * time.Hour)
	apiKey := &APIKey{
		ID:          "test-id",
		Name:        "Test Key",
		Key:         "test-key-value",
		Description: "Test Description",
		Expires:     expiryTime.Format(time.RFC3339),
		Created:     "2023-01-01T00:00:00Z",
		Updated:     "2023-01-01T01:00:00Z",
	}

	// Validate fields
	if apiKey.ID != "test-id" {
		t.Errorf("Expected ID to be 'test-id', got '%s'", apiKey.ID)
	}

	if apiKey.Name != "Test Key" {
		t.Errorf("Expected Name to be 'Test Key', got '%s'", apiKey.Name)
	}

	if apiKey.Key != "test-key-value" {
		t.Errorf("Expected Key to be 'test-key-value', got '%s'", apiKey.Key)
	}

	if apiKey.Description != "Test Description" {
		t.Errorf("Expected Description to be 'Test Description', got '%s'", apiKey.Description)
	}

	if apiKey.Created != "2023-01-01T00:00:00Z" {
		t.Errorf("Expected Created to be '2023-01-01T00:00:00Z', got '%s'", apiKey.Created)
	}

	if apiKey.Updated != "2023-01-01T01:00:00Z" {
		t.Errorf("Expected Updated to be '2023-01-01T01:00:00Z', got '%s'", apiKey.Updated)
	}
}

func TestAPIKeyBinding(t *testing.T) {
	// Test creating an APIKeyBinding instance
	binding := &APIKeyBinding{
		ID:           "binding-id",
		APIKeyID:     "key-id",
		ResourceType: "user",
		ResourceIDs:  []string{"user-1", "user-2"},
		Permissions:  []string{"read", "write"},
		Created:      "2023-01-01T00:00:00Z",
		Updated:      "2023-01-01T01:00:00Z",
	}

	// Validate fields
	if binding.ID != "binding-id" {
		t.Errorf("Expected ID to be 'binding-id', got '%s'", binding.ID)
	}

	if binding.APIKeyID != "key-id" {
		t.Errorf("Expected APIKeyID to be 'key-id', got '%s'", binding.APIKeyID)
	}

	if binding.ResourceType != "user" {
		t.Errorf("Expected ResourceType to be 'user', got '%s'", binding.ResourceType)
	}

	if len(binding.ResourceIDs) != 2 {
		t.Errorf("Expected ResourceIDs to have 2 items, got %d", len(binding.ResourceIDs))
	}

	if binding.ResourceIDs[0] != "user-1" {
		t.Errorf("Expected ResourceIDs[0] to be 'user-1', got '%s'", binding.ResourceIDs[0])
	}

	if binding.ResourceIDs[1] != "user-2" {
		t.Errorf("Expected ResourceIDs[1] to be 'user-2', got '%s'", binding.ResourceIDs[1])
	}

	if len(binding.Permissions) != 2 {
		t.Errorf("Expected Permissions to have 2 items, got %d", len(binding.Permissions))
	}

	if binding.Permissions[0] != "read" {
		t.Errorf("Expected Permissions[0] to be 'read', got '%s'", binding.Permissions[0])
	}

	if binding.Permissions[1] != "write" {
		t.Errorf("Expected Permissions[1] to be 'write', got '%s'", binding.Permissions[1])
	}

	if binding.Created != "2023-01-01T00:00:00Z" {
		t.Errorf("Expected Created to be '2023-01-01T00:00:00Z', got '%s'", binding.Created)
	}

	if binding.Updated != "2023-01-01T01:00:00Z" {
		t.Errorf("Expected Updated to be '2023-01-01T01:00:00Z', got '%s'", binding.Updated)
	}
}
