package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// providerFactories are used to instantiate a provider during testing
var providerFactories = map[string]func() (*schema.Provider, error){
	"jumpcloud": func() (*schema.Provider, error) {
		return New(), nil
	},
}

func TestProvider(t *testing.T) {
	if err := New().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

// testAccPreCheck checks if the required environment variables are set
// and skips the test if they aren't
func testAccPreCheck(t *testing.T) bool {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
		return false
	}

	if v := os.Getenv("JUMPCLOUD_API_KEY"); v == "" {
		t.Skip("JUMPCLOUD_API_KEY must be set for acceptance tests")
		return false
	}

	return true
}

// testAccProvider is used to instantiate a provider during tests
func testAccProvider() *schema.Provider {
	return New()
}

// testAccProviderFactories returns the provider factories for testing
func testAccProviderFactories() map[string]func() (*schema.Provider, error) {
	return providerFactories
}
