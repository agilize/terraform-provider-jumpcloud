package apiclient

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestDoRequestWithContextUsesBearerToken verifies that when OAuth2
// service-account credentials are configured, the client performs the
// client_credentials flow and sends an Authorization: Bearer header on API
// requests (instead of x-api-key).
func TestDoRequestWithContextUsesBearerToken(t *testing.T) {
	const (
		clientID     = "test-client-id"
		clientSecret = "test-client-secret"
		accessToken  = "test-access-token"
	)

	var tokenRequests int

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenRequests++

		wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret))
		if got := r.Header.Get("Authorization"); got != wantAuth {
			t.Errorf("token request Authorization = %q, want %q", got, wantAuth)
		}
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Errorf("token request Content-Type = %q, want application/x-www-form-urlencoded", got)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parsing token request form: %v", err)
		}
		if got := r.Form.Get("grant_type"); got != "client_credentials" {
			t.Errorf("token request grant_type = %q, want client_credentials", got)
		}
		if got := r.Form.Get("scope"); got != "api" {
			t.Errorf("token request scope = %q, want api", got)
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"access_token":"` + accessToken + `","expires_in":3600,"token_type":"Bearer"}`)); err != nil {
			t.Errorf("writing token response: %v", err)
		}
	}))
	defer tokenServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer "+accessToken {
			t.Errorf("api request Authorization = %q, want Bearer %s", got, accessToken)
		}
		if got := r.Header.Get("x-api-key"); got != "" {
			t.Errorf("api request unexpectedly sent x-api-key = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"success": true}`)); err != nil {
			t.Errorf("writing api response: %v", err)
		}
	}))
	defer apiServer.Close()

	client := NewClient(&Config{
		APIURL:        apiServer.URL,
		ClientID:      clientID,
		ClientSecret:  clientSecret,
		OAuthTokenURL: tokenServer.URL,
	})

	ctx := context.Background()

	if _, err := client.DoRequestWithContext(ctx, http.MethodGet, "/test", nil); err != nil {
		t.Fatalf("first request error = %v", err)
	}
	// A second request should reuse the cached token (no new token fetch).
	if _, err := client.DoRequestWithContext(ctx, http.MethodGet, "/test", nil); err != nil {
		t.Fatalf("second request error = %v", err)
	}

	if tokenRequests != 1 {
		t.Errorf("token endpoint hit %d times, want 1 (token should be cached)", tokenRequests)
	}
}

// TestDoRequestWithContextFallsBackToAPIKey verifies that without OAuth2
// credentials the client still sends the static x-api-key header.
func TestDoRequestWithContextFallsBackToAPIKey(t *testing.T) {
	const apiKey = "static-api-key"

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("x-api-key"); got != apiKey {
			t.Errorf("api request x-api-key = %q, want %q", got, apiKey)
		}
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("api request unexpectedly sent Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"success": true}`)); err != nil {
			t.Errorf("writing api response: %v", err)
		}
	}))
	defer apiServer.Close()

	client := NewClient(&Config{
		APIURL: apiServer.URL,
		APIKey: apiKey,
	})

	if _, err := client.DoRequestWithContext(context.Background(), http.MethodGet, "/test", nil); err != nil {
		t.Fatalf("request error = %v", err)
	}
}
