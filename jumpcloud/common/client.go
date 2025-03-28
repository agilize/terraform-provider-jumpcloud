package common

import (
	"context"
	"fmt"
	"net/http"
)

// APIClientInterface defines the interface for the JumpCloud API client
type APIClientInterface interface {
	DoRequest(method, path string, body interface{}) ([]byte, error)
	GetAPIKey() string
	GetOrgID() string
}

// ConvertToClientInterface converts a meta interface to an APIClientInterface
func ConvertToClientInterface(meta interface{}) (APIClientInterface, error) {
	client, ok := meta.(APIClientInterface)
	if !ok {
		return nil, fmt.Errorf("invalid client type")
	}
	return client, nil
}

// IsNotFound checks if the error is a 404 Not Found error
func IsNotFound(statusCode int) bool {
	return statusCode == http.StatusNotFound
}

// ClientContext holds the context for API requests
type ClientContext struct {
	Context context.Context
	Client  APIClientInterface
}

// NewClientContext creates a new ClientContext
func NewClientContext(ctx context.Context, meta interface{}) (*ClientContext, error) {
	client, err := ConvertToClientInterface(meta)
	if err != nil {
		return nil, err
	}
	return &ClientContext{
		Context: ctx,
		Client:  client,
	}, nil
}
