package apiclient

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// CreateMockServer creates a mock HTTP server for testing
func CreateMockServer(t *testing.T, method string) (*httptest.Server, *Client) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request has a context
		if r.Context() == nil {
			t.Error("Request context is nil")
		}

		// Check if the request method is correct
		if r.Method != method {
			t.Errorf("Expected method %s, got %s", method, r.Method)
		}

		// Check if the request has the expected headers
		if r.Header.Get("x-api-key") != "test-api-key" {
			t.Errorf("Expected x-api-key header to be 'test-api-key', got %s", r.Header.Get("x-api-key"))
		}

		// Return a successful response
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"success": true}`)); err != nil {
			t.Errorf("Error writing response: %v", err)
		}
	}))

	// Create a client with the test server URL
	config := &Config{
		APIKey:         "test-api-key",
		OrgID:          "test-org-id",
		APIURL:         server.URL,
		RequestTimeout: 0, // No timeout for tests
	}
	client := NewClient(config)

	return server, client
}
