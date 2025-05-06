package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// JUMPCLOUD_API_V1_URL is the base URL for JumpCloud API v1
const JUMPCLOUD_API_V1_URL = "https://console.jumpcloud.com"

// JUMPCLOUD_API_V2_URL is the base URL for JumpCloud API v2
const JUMPCLOUD_API_V2_URL = "https://console.jumpcloud.com"

// APIVersion represents the JumpCloud API version
type APIVersion string

// Available JumpCloud API versions
const (
	V1 APIVersion = "v1"
	V2 APIVersion = "v2"
)

// Config contains the configuration for the JumpCloud client
// For API documentation, see: https://docs.jumpcloud.com/api/
type Config struct {
	// APIKey is the authentication key for the JumpCloud API
	// See: https://docs.jumpcloud.com/api/authentication
	APIKey string

	// OrgID is the organization ID for multi-tenant operations
	// Required for some API operations in multi-tenant environments
	OrgID string

	// APIURL is the base URL for the JumpCloud API
	// Defaults to https://console.jumpcloud.com/api
	APIURL string

	// Version specifies which API version to use (v1 or v2)
	// Defaults to v2 which is recommended for most operations
	Version APIVersion

	// RequestTimeout is the timeout for API requests
	// Defaults to 30 seconds
	RequestTimeout time.Duration
}

// Client is used to communicate with the JumpCloud API
// It handles authentication, request formatting, and response processing
type Client struct {
	// APIKey is the authentication key for the JumpCloud API
	APIKey string

	// OrgID is the organization ID for multi-tenant operations
	OrgID string

	// APIURL is the base URL for the JumpCloud API
	APIURL string

	// Version specifies which API version to use
	Version APIVersion

	// HTTPClient is the underlying HTTP client used for API requests
	HTTPClient *http.Client
}

// NewClient creates a new JumpCloud client with the provided configuration
// It sets default values for any configuration options that were not specified
func NewClient(config *Config) *Client {
	// Set default timeout if not specified
	timeout := config.RequestTimeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Set default API URL if not specified
	apiURL := config.APIURL
	if apiURL == "" {
		apiURL = JUMPCLOUD_API_V1_URL + "/api"
	}

	// Set default API version if not specified
	version := config.Version
	if version == "" {
		version = V2
	}

	return &Client{
		APIKey:     config.APIKey,
		OrgID:      config.OrgID,
		APIURL:     apiURL,
		Version:    version,
		HTTPClient: &http.Client{Timeout: timeout},
	}
}

// DoRequestWithContext makes an HTTP request to the JumpCloud API with context
// It handles authentication, error handling, and response processing
//
// Parameters:
// - ctx: Context for the request (can be used for cancellation and timeouts)
// - method: HTTP method (GET, POST, PUT, DELETE)
// - path: API endpoint path (e.g. "/systemusers")
// - body: Request body to be serialized as JSON (can be nil)
//
// Returns:
// - Response body as a byte array
// - Error if the request failed
//
// API documentation: https://docs.jumpcloud.com/api/
func (c *Client) DoRequestWithContext(ctx context.Context, method, path string, body any) ([]byte, error) {
	var reqBody io.Reader

	// Convert body to JSON if provided
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshalling request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	// Construct full URL
	url := fmt.Sprintf("%s%s", c.APIURL, path)

	// Log the full URL for debugging
	fmt.Printf("DEBUG: Making request to URL: %s\n", url)

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	// JumpCloud API requires x-api-key header for authentication
	// See: https://docs.jumpcloud.com/api/authentication
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add organization ID if provided
	// Required for multi-tenant operations
	if c.OrgID != "" {
		req.Header.Set("x-org-id", c.OrgID)
	}

	// Execute the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Log response status and body for debugging
	fmt.Printf("DEBUG: Response status: %d\n", resp.StatusCode)
	fmt.Printf("DEBUG: Response body: %s\n", string(respBody))

	// Check for HTTP error status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, ParseJumpCloudError(resp.StatusCode, respBody)
	}

	return respBody, nil
}

// DoRequest makes an HTTP request to the JumpCloud API
// It handles authentication, error handling, and response processing
//
// Parameters:
// - method: HTTP method (GET, POST, PUT, DELETE)
// - path: API endpoint path (e.g. "/systemusers")
// - body: Request body to be serialized as JSON (can be nil)
//
// Returns:
// - Response body as a byte array
// - Error if the request failed
//
// API documentation: https://docs.jumpcloud.com/api/
func (c *Client) DoRequest(method, path string, body any) ([]byte, error) {
	// Use the context-aware version with a background context
	return c.DoRequestWithContext(context.Background(), method, path, body)
}

// GetV1WithContext is a convenience method for making GET requests to the JumpCloud API v1 with context
func (c *Client) GetV1WithContext(ctx context.Context, path string) ([]byte, error) {
	tempAPIURL := c.APIURL
	c.APIURL = JUMPCLOUD_API_V1_URL
	defer func() { c.APIURL = tempAPIURL }()

	// Ensure path starts with /api
	if !strings.HasPrefix(path, "/api") {
		path = "/api" + path
	}

	return c.DoRequestWithContext(ctx, http.MethodGet, path, nil)
}

// GetV1 is a convenience method for making GET requests to the JumpCloud API v1
func (c *Client) GetV1(path string) ([]byte, error) {
	return c.GetV1WithContext(context.Background(), path)
}

// GetV2WithContext is a convenience method for making GET requests to the JumpCloud API v2 with context
func (c *Client) GetV2WithContext(ctx context.Context, path string) ([]byte, error) {
	tempAPIURL := c.APIURL
	c.APIURL = JUMPCLOUD_API_V2_URL
	defer func() { c.APIURL = tempAPIURL }()

	// Ensure path starts with /api/v2
	if !strings.HasPrefix(path, "/api/v2") {
		if strings.HasPrefix(path, "/") {
			path = "/api/v2" + path
		} else {
			path = "/api/v2/" + path
		}
	}

	return c.DoRequestWithContext(ctx, http.MethodGet, path, nil)
}

// GetV2 is a convenience method for making GET requests to the JumpCloud API v2
func (c *Client) GetV2(path string) ([]byte, error) {
	return c.GetV2WithContext(context.Background(), path)
}

// PostV1WithContext is a convenience method for making POST requests to the JumpCloud API v1 with context
func (c *Client) PostV1WithContext(ctx context.Context, path string, body any) ([]byte, error) {
	tempAPIURL := c.APIURL
	c.APIURL = JUMPCLOUD_API_V1_URL
	defer func() { c.APIURL = tempAPIURL }()

	// Ensure path starts with /api
	if !strings.HasPrefix(path, "/api") {
		if strings.HasPrefix(path, "/") {
			path = "/api" + path
		} else {
			path = "/api/" + path
		}
	}

	return c.DoRequestWithContext(ctx, http.MethodPost, path, body)
}

// PostV2WithContext is a convenience method for making POST requests to the JumpCloud API v2 with context
func (c *Client) PostV2WithContext(ctx context.Context, path string, body any) ([]byte, error) {
	tempAPIURL := c.APIURL
	c.APIURL = JUMPCLOUD_API_V2_URL
	defer func() { c.APIURL = tempAPIURL }()

	// Ensure path starts with /api/v2
	if !strings.HasPrefix(path, "/api/v2") {
		if strings.HasPrefix(path, "/") {
			path = "/api/v2" + path
		} else {
			path = "/api/v2/" + path
		}
	}

	return c.DoRequestWithContext(ctx, http.MethodPost, path, body)
}

// PutV1WithContext is a convenience method for making PUT requests to the JumpCloud API v1 with context
func (c *Client) PutV1WithContext(ctx context.Context, path string, body any) ([]byte, error) {
	tempAPIURL := c.APIURL
	c.APIURL = JUMPCLOUD_API_V1_URL
	defer func() { c.APIURL = tempAPIURL }()

	// Ensure path starts with /api
	if !strings.HasPrefix(path, "/api") {
		if strings.HasPrefix(path, "/") {
			path = "/api" + path
		} else {
			path = "/api/" + path
		}
	}

	return c.DoRequestWithContext(ctx, http.MethodPut, path, body)
}

// PutV2WithContext is a convenience method for making PUT requests to the JumpCloud API v2 with context
func (c *Client) PutV2WithContext(ctx context.Context, path string, body any) ([]byte, error) {
	tempAPIURL := c.APIURL
	c.APIURL = JUMPCLOUD_API_V2_URL
	defer func() { c.APIURL = tempAPIURL }()

	// Ensure path starts with /api/v2
	if !strings.HasPrefix(path, "/api/v2") {
		if strings.HasPrefix(path, "/") {
			path = "/api/v2" + path
		} else {
			path = "/api/v2/" + path
		}
	}

	return c.DoRequestWithContext(ctx, http.MethodPut, path, body)
}

// DeleteV1WithContext is a convenience method for making DELETE requests to the JumpCloud API v1 with context
func (c *Client) DeleteV1WithContext(ctx context.Context, path string) ([]byte, error) {
	tempAPIURL := c.APIURL
	c.APIURL = JUMPCLOUD_API_V1_URL
	defer func() { c.APIURL = tempAPIURL }()

	// Ensure path starts with /api
	if !strings.HasPrefix(path, "/api") {
		if strings.HasPrefix(path, "/") {
			path = "/api" + path
		} else {
			path = "/api/" + path
		}
	}

	return c.DoRequestWithContext(ctx, http.MethodDelete, path, nil)
}

// DeleteV2WithContext is a convenience method for making DELETE requests to the JumpCloud API v2 with context
func (c *Client) DeleteV2WithContext(ctx context.Context, path string) ([]byte, error) {
	tempAPIURL := c.APIURL
	c.APIURL = JUMPCLOUD_API_V2_URL
	defer func() { c.APIURL = tempAPIURL }()

	// Ensure path starts with /api/v2
	if !strings.HasPrefix(path, "/api/v2") {
		if strings.HasPrefix(path, "/") {
			path = "/api/v2" + path
		} else {
			path = "/api/v2/" + path
		}
	}

	return c.DoRequestWithContext(ctx, http.MethodDelete, path, nil)
}

// PostV1 is a convenience method for making POST requests to the JumpCloud API v1
func (c *Client) PostV1(path string, body any) ([]byte, error) {
	return c.PostV1WithContext(context.Background(), path, body)
}

// PostV2 is a convenience method for making POST requests to the JumpCloud API v2
func (c *Client) PostV2(path string, body any) ([]byte, error) {
	return c.PostV2WithContext(context.Background(), path, body)
}

// PutV1 is a convenience method for making PUT requests to the JumpCloud API v1
func (c *Client) PutV1(path string, body any) ([]byte, error) {
	return c.PutV1WithContext(context.Background(), path, body)
}

// PutV2 is a convenience method for making PUT requests to the JumpCloud API v2
func (c *Client) PutV2(path string, body any) ([]byte, error) {
	return c.PutV2WithContext(context.Background(), path, body)
}

// DeleteV1 is a convenience method for making DELETE requests to the JumpCloud API v1
func (c *Client) DeleteV1(path string) ([]byte, error) {
	return c.DeleteV1WithContext(context.Background(), path)
}

// DeleteV2 is a convenience method for making DELETE requests to the JumpCloud API v2
func (c *Client) DeleteV2(path string) ([]byte, error) {
	return c.DeleteV2WithContext(context.Background(), path)
}

// GetApiKey returns the API key used by the client
func (c *Client) GetApiKey() string {
	return c.APIKey
}

// GetOrgID returns the organization ID used by the client
func (c *Client) GetOrgID() string {
	return c.OrgID
}

// IsResourceNotFound checks if the error is a "not found" (404) error
func (c *Client) IsResourceNotFound(err error) bool {
	if err == nil {
		return false
	}

	jumpCloudErr, ok := err.(*JumpCloudError)
	if !ok {
		return false
	}

	return jumpCloudErr.StatusCode == http.StatusNotFound
}

// IsResourceAlreadyExists checks if the error is a "conflict" (409) error
// indicating that a resource with the same identifier already exists
func (c *Client) IsResourceAlreadyExists(err error) bool {
	if err == nil {
		return false
	}

	jumpCloudErr, ok := err.(*JumpCloudError)
	if !ok {
		return false
	}

	return jumpCloudErr.StatusCode == http.StatusConflict
}

// IsPermissionDenied checks if the error is a permission denied (403) error
func (c *Client) IsPermissionDenied(err error) bool {
	if err == nil {
		return false
	}

	jumpCloudErr, ok := err.(*JumpCloudError)
	if !ok {
		return false
	}

	return jumpCloudErr.StatusCode == http.StatusForbidden
}

// IsAuthenticationError checks if the error is an authentication error (401)
func (c *Client) IsAuthenticationError(err error) bool {
	if err == nil {
		return false
	}

	jumpCloudErr, ok := err.(*JumpCloudError)
	if !ok {
		return false
	}

	return jumpCloudErr.StatusCode == http.StatusUnauthorized
}

// IsInvalidInput checks if the error is a bad request error (400)
func (c *Client) IsInvalidInput(err error) bool {
	if err == nil {
		return false
	}

	jumpCloudErr, ok := err.(*JumpCloudError)
	if !ok {
		return false
	}

	return jumpCloudErr.StatusCode == http.StatusBadRequest
}
