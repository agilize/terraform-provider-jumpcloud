package common

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// ClientInterface defines the methods a JumpCloud client must implement
type ClientInterface interface {
	// DoRequest performs an API request with the given method, path, and body
	// The body parameter accepts interface{} to support both []byte and struct types
	DoRequest(method, path string, body interface{}) ([]byte, error)

	// DoRequestWithContext performs an API request with context and the given method, path, and body
	// The body parameter accepts interface{} to support both []byte and struct types
	DoRequestWithContext(ctx context.Context, method, path string, body interface{}) ([]byte, error)

	// GetApiKey returns the API key used for authentication
	GetApiKey() string

	// GetOrgID returns the organization ID
	GetOrgID() string
}

// GetClientFromMeta converts the meta interface to a ClientInterface
func GetClientFromMeta(meta any) (ClientInterface, diag.Diagnostics) {
	if meta == nil {
		return nil, diag.Errorf("meta value is nil")
	}

	client, ok := meta.(ClientInterface)
	if !ok {
		return nil, diag.Errorf("invalid client type: %T", meta)
	}

	return client, nil
}
