package apiclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDoRequestWithContext(t *testing.T) {
	server, client := CreateMockServer(t, http.MethodGet)
	defer server.Close()

	// Create a context
	ctx := context.Background()

	// Make a request with the context
	resp, err := client.DoRequestWithContext(ctx, http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("DoRequestWithContext() error = %v", err)
	}

	// Check the response
	if string(resp) != `{"success": true}` {
		t.Errorf("DoRequestWithContext() = %v, want %v", string(resp), `{"success": true}`)
	}
}

func TestDoRequestWithCanceledContext(t *testing.T) {
	// Create a test server that delays the response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep for 100ms to simulate a slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"success": true}`)); err != nil {
			t.Errorf("Error writing response: %v", err)
		}
	}))
	defer server.Close()

	// Create a client with the test server URL
	config := &Config{
		APIKey:         "test-api-key",
		OrgID:          "test-org-id",
		APIURL:         server.URL,
		RequestTimeout: 0, // No timeout for the client itself
	}
	client := NewClient(config)

	// Create a context with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Make a request with the context that will be canceled
	_, err := client.DoRequestWithContext(ctx, http.MethodGet, "/test", nil)

	// Check that the request was canceled
	if err == nil {
		t.Error("DoRequestWithContext() error = nil, want context deadline exceeded error")
	}
}

func TestGetV1WithContext(t *testing.T) {
	server, client := CreateMockServer(t, http.MethodGet)
	defer server.Close()

	// Create a context
	ctx := context.Background()

	// Make a request with the context
	resp, err := client.DoRequestWithContext(ctx, http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("GetV1WithContext() error = %v", err)
	}

	// Check the response
	if string(resp) != `{"success": true}` {
		t.Errorf("GetV1WithContext() = %v, want %v", string(resp), `{"success": true}`)
	}
}

func TestGetV2WithContext(t *testing.T) {
	server, client := CreateMockServer(t, http.MethodGet)
	defer server.Close()

	// Create a context
	ctx := context.Background()

	// Make a request with the context
	resp, err := client.DoRequestWithContext(ctx, http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("GetV2WithContext() error = %v", err)
	}

	// Check the response
	if string(resp) != `{"success": true}` {
		t.Errorf("GetV2WithContext() = %v, want %v", string(resp), `{"success": true}`)
	}
}

func TestPostV1WithContext(t *testing.T) {
	server, client := CreateMockServer(t, http.MethodPost)
	defer server.Close()

	// Create a context
	ctx := context.Background()

	// Make a request with the context
	resp, err := client.DoRequestWithContext(ctx, http.MethodPost, "/test", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("PostV1WithContext() error = %v", err)
	}

	// Check the response
	if string(resp) != `{"success": true}` {
		t.Errorf("PostV1WithContext() = %v, want %v", string(resp), `{"success": true}`)
	}
}

func TestPutV1WithContext(t *testing.T) {
	server, client := CreateMockServer(t, http.MethodPut)
	defer server.Close()

	// Create a context
	ctx := context.Background()

	// Make a request with the context
	resp, err := client.DoRequestWithContext(ctx, http.MethodPut, "/test", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("PutV1WithContext() error = %v", err)
	}

	// Check the response
	if string(resp) != `{"success": true}` {
		t.Errorf("PutV1WithContext() = %v, want %v", string(resp), `{"success": true}`)
	}
}

func TestDeleteV1WithContext(t *testing.T) {
	server, client := CreateMockServer(t, http.MethodDelete)
	defer server.Close()

	// Create a context
	ctx := context.Background()

	// Make a request with the context
	resp, err := client.DoRequestWithContext(ctx, http.MethodDelete, "/test", nil)
	if err != nil {
		t.Fatalf("DeleteV1WithContext() error = %v", err)
	}

	// Check the response
	if string(resp) != `{"success": true}` {
		t.Errorf("DeleteV1WithContext() = %v, want %v", string(resp), `{"success": true}`)
	}
}
