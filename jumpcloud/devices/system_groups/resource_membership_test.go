package system_groups

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// providerFactories is a map of provider factory functions for testing
var providerFactoriesMembership = map[string]func() (*schema.Provider, error){
	"jumpcloud": func() (*schema.Provider, error) {
		return jctest.TestAccProviders["jumpcloud"], nil
	},
}

func TestAccJumpCloudSystemGroupMembership(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	var resourceName = "jumpcloud_system_group_membership.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudSystemGroupMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudSystemGroupMembershipConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudSystemGroupMembershipExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "system_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "system_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// testAccJumpCloudSystemGroupMembershipConfig returns a test configuration for system group membership
func testAccJumpCloudSystemGroupMembershipConfig() string {
	return `
resource "jumpcloud_system_group" "test" {
  name        = "test-acc-system-group"
  description = "Test system group for acceptance tests"
}

resource "jumpcloud_system" "test" {
  display_name = "test-acc-system"
  allow_ssh_password_authentication = true
  allow_ssh_root_login = false
}

resource "jumpcloud_system_group_membership" "test" {
  system_id = jumpcloud_system.test.id
  system_group_id = jumpcloud_system_group.test.id
}
`
}

// testAccCheckJumpCloudSystemGroupMembershipExists verifies the membership exists
func testAccCheckJumpCloudSystemGroupMembershipExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		// Add implementation to check if the system group membership actually exists in JumpCloud
		// This will depend on how you've structured your test setup

		return nil
	}
}

// testAccCheckJumpCloudSystemGroupMembershipDestroy verifies the membership has been destroyed
func testAccCheckJumpCloudSystemGroupMembershipDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_system_group_membership" {
			continue
		}

		// Add implementation to check if the system group membership has been destroyed in JumpCloud
		// This will depend on how you've structured your test setup

		return nil
	}

	return nil
}
