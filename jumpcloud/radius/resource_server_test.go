package radius

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccJumpCloudRadiusServer(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	var resourceName = "jumpcloud_radius_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckJumpCloudRadiusServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudRadiusServerConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudRadiusServerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test-radius-server"),
					resource.TestCheckResourceAttr(resourceName, "network_source_ip", "192.168.1.1"),
					resource.TestCheckResourceAttr(resourceName, "mfa_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "user_password_expiration_action", "allow"),
					resource.TestCheckResourceAttr(resourceName, "user_lockout_action", "deny"),
					resource.TestCheckResourceAttr(resourceName, "user_attribute", "username"),
				),
			},
			{
				Config: testAccJumpCloudRadiusServerConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudRadiusServerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test-radius-server-updated"),
					resource.TestCheckResourceAttr(resourceName, "network_source_ip", "192.168.1.2"),
					resource.TestCheckResourceAttr(resourceName, "mfa_required", "true"),
					resource.TestCheckResourceAttr(resourceName, "user_password_expiration_action", "deny"),
					resource.TestCheckResourceAttr(resourceName, "user_lockout_action", "allow"),
					resource.TestCheckResourceAttr(resourceName, "user_attribute", "email"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore shared_secret as it's not returned by the API for security reasons
				ImportStateVerifyIgnore: []string{"shared_secret"},
			},
		},
	})
}

func testAccJumpCloudRadiusServerConfig() string {
	return `
resource "jumpcloud_radius_server" "test" {
  name                            = "test-radius-server"
  shared_secret                   = "secretT3stP@ss!"
  network_source_ip               = "192.168.1.1"
  mfa_required                    = false
  user_password_expiration_action = "allow"
  user_lockout_action             = "deny"
  user_attribute                  = "username"
}
`
}

func testAccJumpCloudRadiusServerConfigUpdated() string {
	return `
resource "jumpcloud_radius_server" "test" {
  name                            = "test-radius-server-updated"
  shared_secret                   = "updatedSecretT3stP@ss!"
  network_source_ip               = "192.168.1.2"
  mfa_required                    = true
  user_password_expiration_action = "deny"
  user_lockout_action             = "allow"
  user_attribute                  = "email"
}
`
}

func testAccCheckJumpCloudRadiusServerExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		// Add implementation to check if the RADIUS server actually exists in JumpCloud
		// This will depend on how you've structured your test setup

		return nil
	}
}

func testAccCheckJumpCloudRadiusServerDestroy(s *terraform.State) error {
	// This function will be called at the end of the test to ensure
	// all resources have been properly cleaned up
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_radius_server" {
			continue
		}

		// Add implementation to check the resource has been deleted from JumpCloud
		// This will depend on how you've structured your test setup
	}

	return nil
}
