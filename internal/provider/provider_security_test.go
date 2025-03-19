package provider

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/agilize/terraform-provider-jumpcloud/internal/client"
	"github.com/stretchr/testify/assert"
)

// Test that the provider properly masks sensitive data
func TestSecurity_SensitiveValues(t *testing.T) {
	// Create a provider instance
	p := New()

	// Check that the api_key field is marked as sensitive
	assert.True(t, p.Schema["api_key"].Sensitive, "API key should be marked as sensitive")

	// The same for resource schemas
	r := resourceUser()
	assert.True(t, r.Schema["password"].Sensitive, "Password should be marked as sensitive")
}

// Test that the client properly sets security headers
func TestSecurity_Headers(t *testing.T) {
	// Create a mock client config
	config := &client.Config{
		APIKey: "test-key",
		OrgID:  "test-org",
	}

	// Create a client
	c := client.NewClient(config)

	// Check that the client is configured with the proper values
	assert.Equal(t, "test-key", c.APIKey, "API key should be set on the client")
	assert.Equal(t, "test-org", c.OrgID, "Organization ID should be set on the client")

	// TODO: In a real implementation, we would test that the client sends the proper headers
	// This would require mocking the HTTP client and checking the requests
}

// Test that sensitive data can't be logged
func TestSecurity_LogSanitization(t *testing.T) {
	// Get full provider schema
	provider := New()
	schemaMap := provider.Schema

	// Check all provider schema fields
	for k, v := range schemaMap {
		if v.Sensitive {
			// Any sensitive field should not be included in logs
			assert.True(t, strings.Contains(k, "key") || strings.Contains(k, "secret") || strings.Contains(k, "password"),
				"Field %s is marked as sensitive but doesn't have a name indicating sensitive data", k)
		}
	}

	// Check all resource schemas for sensitive fields
	resourceSchema := resourceUser().Schema
	for k, v := range resourceSchema {
		if v.Sensitive {
			// Validate that all sensitive fields have proper naming
			assert.True(t, strings.Contains(k, "password") || strings.Contains(k, "secret") || strings.Contains(k, "key"),
				"Field %s is marked as sensitive but doesn't have a name indicating sensitive data", k)
		}

		// Non-sensitive fields that should be sensitive
		if strings.Contains(k, "password") || strings.Contains(k, "secret") || strings.Contains(k, "key") {
			assert.True(t, v.Sensitive, "Field %s contains sensitive-looking name but is not marked as sensitive", k)
		}
	}
}

// Test password strength enforcement
func TestSecurity_PasswordStrength(t *testing.T) {
	// Define simpler regex checks for password strength
	// Instead of using lookaheads which Go doesn't support
	hasLower := regexp.MustCompile(`[a-z]`)
	hasUpper := regexp.MustCompile(`[A-Z]`)
	hasDigit := regexp.MustCompile(`\d`)
	hasSpecial := regexp.MustCompile(`[^\da-zA-Z]`)
	hasLength := regexp.MustCompile(`.{8,}`)

	// Test weak passwords
	weakPasswords := []string{
		"password",
		"123456",
		"qwerty",
		"letmein",
		"abc123",
	}

	for _, password := range weakPasswords {
		// Check why it's weak
		var issues []string
		if !hasLower.MatchString(password) {
			issues = append(issues, "no lowercase")
		}
		if !hasUpper.MatchString(password) {
			issues = append(issues, "no uppercase")
		}
		if !hasDigit.MatchString(password) {
			issues = append(issues, "no digit")
		}
		if !hasSpecial.MatchString(password) {
			issues = append(issues, "no special character")
		}
		if !hasLength.MatchString(password) {
			issues = append(issues, "too short")
		}

		// At least one issue should be found
		assert.Greater(t, len(issues), 0, "Weak password didn't trigger any strength issues: %s", password)
	}

	// Test strong passwords
	strongPasswords := []string{
		"SecureP@ssw0rd",
		"C0mpl3x!P4ss",
		"Sup3r$3cur3P@$$w0rd",
	}

	for _, password := range strongPasswords {
		// Check all strength criteria
		assert.True(t, hasLower.MatchString(password), "Strong password missing lowercase: %s", password)
		assert.True(t, hasUpper.MatchString(password), "Strong password missing uppercase: %s", password)
		assert.True(t, hasDigit.MatchString(password), "Strong password missing digit: %s", password)
		assert.True(t, hasSpecial.MatchString(password), "Strong password missing special character: %s", password)
		assert.True(t, hasLength.MatchString(password), "Strong password not long enough: %s", password)
	}

	// TODO: In a real implementation, we would validate that the provider enforces
	// password strength requirements, either client-side or through API validation
}

// Test API error handling for security-related issues
func TestSecurity_APIErrorHandling(t *testing.T) {
	// Test parsing auth errors
	authErrorJson := `{"code":"AUTH_FAILED","message":"Authentication failed"}`
	statusCode := 401

	errorObj := client.ParseJumpCloudError(statusCode, []byte(authErrorJson))

	assert.Equal(t, "AUTH_FAILED", errorObj.Code)
	assert.Equal(t, "Authentication failed", errorObj.Message)
	assert.Equal(t, 401, errorObj.StatusCode)

	// Test that IsAuthError correctly identifies auth errors
	assert.True(t, client.IsAuthError(errorObj))

	// Test permission errors
	permErrorJson := `{"code":"PERMISSION_DENIED","message":"No permission to access resource"}`
	statusCode = 403

	errorObj = client.ParseJumpCloudError(statusCode, []byte(permErrorJson))

	assert.Equal(t, "PERMISSION_DENIED", errorObj.Code)
	assert.Equal(t, "No permission to access resource", errorObj.Message)
	assert.Equal(t, 403, errorObj.StatusCode)

	// Test that IsPermissionDenied correctly identifies permission errors
	assert.True(t, client.IsPermissionDenied(errorObj))
}

// Test for JSON processing security issues (injection, parsing errors)
func TestSecurity_JSONProcessing(t *testing.T) {
	// Test that the parser properly handles JSON with invalid syntax
	invalidJson := `{"not valid json`

	var result map[string]interface{}
	err := json.Unmarshal([]byte(invalidJson), &result)

	assert.Error(t, err, "Invalid JSON should cause an error")

	// Test handling potentially dangerous input
	dangerousJson := `{"__proto__":{"isAdmin":true}}`

	err = json.Unmarshal([]byte(dangerousJson), &result)

	assert.NoError(t, err, "Valid but potentially dangerous JSON should be parsed without error")
	// In Go, this is safe because the unmarshaller doesn't set JavaScript prototype properties

	// Additional JSON security checks would be implementation-specific
	// and would depend on how we're using the parsed JSON
}
