package common

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// ClientInterface defines the methods a JumpCloud client must implement
type ClientInterface interface {
	// DoRequest performs an API request with the given method, path, and body
	DoRequest(method, path string, body []byte) ([]byte, error)
}

// GetClientFromMeta converts the meta interface to a ClientInterface
func GetClientFromMeta(meta interface{}) (ClientInterface, diag.Diagnostics) {
	if meta == nil {
		return nil, diag.Errorf("meta value is nil")
	}

	client, ok := meta.(ClientInterface)
	if !ok {
		return nil, diag.Errorf("invalid client type: %T", meta)
	}

	return client, nil
}
