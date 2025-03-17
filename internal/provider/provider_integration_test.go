package provider

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ferreirafav/terraform-provider-jumpcloud/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Integration test flag
const (
	INTEGRATION_TEST_ENV = "JUMPCLOUD_INTEGRATION_TEST"
)

// Integration test setup helper
func setupIntegrationTest(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv(INTEGRATION_TEST_ENV) == "" {
		t.Skipf("Skipping integration test, set %s=1 to enable", INTEGRATION_TEST_ENV)
	}

	// Check for required environment variables
	apiKey := os.Getenv("JUMPCLOUD_API_KEY")
	if apiKey == "" {
		t.Fatal("JUMPCLOUD_API_KEY must be set for integration tests")
	}
}

// Integration test for user resource
func TestIntegration_JumpCloudUser(t *testing.T) {
	setupIntegrationTest(t)

	// Generate a unique username to avoid conflicts
	timestamp := time.Now().Unix()
	username := fmt.Sprintf("test-user-%d", timestamp)
	email := fmt.Sprintf("test-user-%d@example.com", timestamp)
	displayName := fmt.Sprintf("Test User %d", timestamp)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testIntegrationUserConfig(username, email, displayName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudUserExists("jumpcloud_user.integration_test"),
					resource.TestCheckResourceAttr("jumpcloud_user.integration_test", "username", username),
					resource.TestCheckResourceAttr("jumpcloud_user.integration_test", "email", email),
					resource.TestCheckResourceAttr("jumpcloud_user.integration_test", "firstname", "Integration"),
					resource.TestCheckResourceAttr("jumpcloud_user.integration_test", "lastname", "Test"),
					resource.TestCheckResourceAttr("jumpcloud_user.integration_test", "description", displayName),
				),
			},
			{
				// Test updating the user
				Config: testIntegrationUserConfigUpdate(username, email, displayName+" Updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudUserExists("jumpcloud_user.integration_test"),
					resource.TestCheckResourceAttr("jumpcloud_user.integration_test", "description", displayName+" Updated"),
					resource.TestCheckResourceAttr("jumpcloud_user.integration_test", "firstname", "Updated"),
				),
			},
		},
	})
}

// Integration test for system resource
func TestIntegration_JumpCloudSystem(t *testing.T) {
	setupIntegrationTest(t)

	// Generate a unique name to avoid conflicts
	timestamp := time.Now().Unix()
	displayName := fmt.Sprintf("test-system-%d", timestamp)
	description := fmt.Sprintf("Test System %d", timestamp)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testIntegrationSystemConfig(displayName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudSystemExists("jumpcloud_system.integration_test"),
					resource.TestCheckResourceAttr("jumpcloud_system.integration_test", "display_name", displayName),
					resource.TestCheckResourceAttr("jumpcloud_system.integration_test", "description", description),
					resource.TestCheckResourceAttr("jumpcloud_system.integration_test", "allow_ssh_root_login", "false"),
				),
			},
			{
				// Test updating the system
				Config: testIntegrationSystemConfigUpdate(displayName, description+" Updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudSystemExists("jumpcloud_system.integration_test"),
					resource.TestCheckResourceAttr("jumpcloud_system.integration_test", "description", description+" Updated"),
					resource.TestCheckResourceAttr("jumpcloud_system.integration_test", "allow_ssh_root_login", "true"),
				),
			},
		},
	})
}

// Create a real client for integration tests
func createIntegrationClient(t *testing.T) *client.Client {
	apiKey := os.Getenv("JUMPCLOUD_API_KEY")
	orgID := os.Getenv("JUMPCLOUD_ORG_ID")

	config := &client.Config{
		APIKey: apiKey,
		OrgID:  orgID,
	}

	return client.NewClient(config)
}

// Integration test configuration for user resource
func testIntegrationUserConfig(username, email, displayName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_user" "integration_test" {
  username    = "%s"
  email       = "%s"
  firstname   = "Integration"
  lastname    = "Test"
  password    = "SecurePassword123!"
  description = "%s"
  
  attributes = {
    integration_test = "true"
    test_timestamp   = "%d"
  }
  
  mfa_enabled            = false
  password_never_expires = true
}
`, username, email, displayName, time.Now().Unix())
}

// Integration test configuration for user resource update
func testIntegrationUserConfigUpdate(username, email, displayName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_user" "integration_test" {
  username    = "%s"
  email       = "%s"
  firstname   = "Updated"
  lastname    = "Test"
  password    = "SecurePassword456!"
  description = "%s"
  
  attributes = {
    integration_test = "true"
    test_timestamp   = "%d"
    updated          = "true"
  }
  
  mfa_enabled            = false
  password_never_expires = true
}
`, username, email, displayName, time.Now().Unix())
}

// Integration test configuration for system resource
func testIntegrationSystemConfig(displayName, description string) string {
	return fmt.Sprintf(`
resource "jumpcloud_system" "integration_test" {
  display_name                      = "%s"
  description                       = "%s"
  allow_ssh_root_login              = false
  allow_ssh_password_authentication = true
  allow_multi_factor_authentication = false
  
  tags = [
    "integration-test",
    "terraform"
  ]
  
  attributes = {
    integration_test = "true"
    test_timestamp   = "%d"
  }
}
`, displayName, description, time.Now().Unix())
}

// Integration test configuration for system resource update
func testIntegrationSystemConfigUpdate(displayName, description string) string {
	return fmt.Sprintf(`
resource "jumpcloud_system" "integration_test" {
  display_name                      = "%s"
  description                       = "%s"
  allow_ssh_root_login              = true
  allow_ssh_password_authentication = true
  allow_multi_factor_authentication = true
  
  tags = [
    "integration-test",
    "terraform",
    "updated"
  ]
  
  attributes = {
    integration_test = "true"
    test_timestamp   = "%d"
    updated          = "true"
  }
}
`, displayName, description, time.Now().Unix())
}
