package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// BenchmarkResourceUserCreate benchmarks user creation
func BenchmarkResourceUserCreate(b *testing.B) {
	r := resourceUser()
	data := r.Data(nil)
	data.Set("username", "benchtest")
	data.Set("email", "benchtest@example.com")
	data.Set("firstname", "Bench")
	data.Set("lastname", "Test")
	data.Set("password", "P@ssw0rd123!")

	// Create a mock provider with a mock client
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_API_KEY", nil),
				Description: "JumpCloud API Key",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_ORG_ID", nil),
				Description: "JumpCloud Organization ID",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"jumpcloud_user": resourceUser(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"jumpcloud_user": dataSourceUser(),
		},
		ConfigureContextFunc: func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
			// Create a mock client for testing
			return &mockClient{}, nil
		},
	}

	meta, _ := p.ConfigureContextFunc(context.Background(), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset the resource ID for each iteration
		data.SetId("")

		// Call CreateContext and track time
		start := time.Now()
		r.CreateContext(context.Background(), data, meta)
		elapsed := time.Since(start)

		// Log the time taken if verbose benchmarking is enabled
		if testing.Verbose() {
			fmt.Printf("Create user iteration %d took %s\n", i, elapsed)
		}
	}
}

// BenchmarkResourceUserRead benchmarks reading user data
func BenchmarkResourceUserRead(b *testing.B) {
	r := resourceUser()
	data := r.Data(nil)
	data.SetId("benchmark-user-id")
	data.Set("username", "benchtest")
	data.Set("email", "benchtest@example.com")

	// Create a mock provider with a mock client
	p := testProvider()
	meta, _ := p.ConfigureContextFunc(context.Background(), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		r.ReadContext(context.Background(), data, meta)
		elapsed := time.Since(start)

		if testing.Verbose() {
			fmt.Printf("Read user iteration %d took %s\n", i, elapsed)
		}
	}
}

// testProvider returns a mock provider for testing
func testProvider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_API_KEY", nil),
				Description: "JumpCloud API Key",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_ORG_ID", nil),
				Description: "JumpCloud Organization ID",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"jumpcloud_user": resourceUser(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"jumpcloud_user": dataSourceUser(),
		},
		ConfigureContextFunc: func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
			// Create a mock client for testing
			return &mockClient{}, nil
		},
	}
}

// TestPerformance_ClientResponseParsing tests the performance of parsing API responses
func TestPerformance_ClientResponseParsing(t *testing.T) {
	// Skip in normal test runs, only run when explicitly testing performance
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Sample user JSON response from JumpCloud API
	userJSON := `{
		"id": "test-user-id",
		"username": "testuser",
		"email": "test@example.com",
		"firstname": "Test",
		"lastname": "User",
		"created": "2023-07-01T12:00:00Z",
		"account_locked": false,
		"activated": true,
		"addresses": [],
		"allow_public_key": true,
		"attributes": [],
		"company": "TestCompany",
		"costCenter": "Test Cost Center",
		"department": "Test Department",
		"description": "Test user description",
		"displayname": "Test User",
		"employeeIdentifier": "E12345",
		"employeeType": "Full-time",
		"enable_managed_uid": false,
		"enable_user_portal_multifactor": true,
		"external_dn": "",
		"external_source_type": "",
		"externally_managed": false,
		"job_title": "Test Job Title",
		"ldap_binding_user": false,
		"location": "Test Location",
		"mfa": {
			"configured": false,
			"exclusion": false
		},
		"middlename": "",
		"password_expiration_date": null,
		"password_expired": false,
		"password_never_expires": false,
		"passwordless_sudo": false,
		"phone_numbers": [],
		"samba_service_user": false,
		"ssh_keys": [],
		"sudo": false,
		"unix_guid": null,
		"unix_uid": null
	}`

	// Time how long it takes to parse 1000 times
	startTime := time.Now()
	iterations := 1000

	for i := 0; i < iterations; i++ {
		var user map[string]interface{}
		err := json.Unmarshal([]byte(userJSON), &user)
		if err != nil {
			t.Fatalf("Failed to parse user JSON: %v", err)
		}

		// Validate a few fields to make sure the parser correctly processed the JSON
		assert.Equal(t, "test-user-id", user["id"])
		assert.Equal(t, "testuser", user["username"])
		assert.Equal(t, "test@example.com", user["email"])
	}

	duration := time.Since(startTime)
	avgTime := duration / time.Duration(iterations)

	t.Logf("Average parse time for user JSON: %s", avgTime)

	// Assert that parsing is reasonably fast (under 50 microseconds per parse)
	assert.Less(t, avgTime, 50*time.Microsecond, "JSON parsing should be fast")
}

// TestPerformance_ResourceSchemaInitialization tests schema initialization performance
func TestPerformance_ResourceSchemaInitialization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	startTime := time.Now()
	iterations := 100

	for i := 0; i < iterations; i++ {
		// Initialize the resource schema
		_ = resourceUser()
		_ = resourceSystem()
		_ = dataSourceUser()
		_ = dataSourceSystem()
	}

	duration := time.Since(startTime)
	avgTime := duration / time.Duration(iterations)

	t.Logf("Average time for resource schema initialization: %s", avgTime)

	// Assert that schema initialization is reasonably fast (under 100 microseconds)
	assert.Less(t, avgTime, 100*time.Microsecond, "Schema initialization should be fast")
}
