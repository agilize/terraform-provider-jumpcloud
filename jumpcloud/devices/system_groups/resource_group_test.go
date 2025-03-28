package system_groups

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// providerFactories is a map of provider factory functions for testing

func TestAccJumpCloudSystemGroup(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	var resourceName = "jumpcloud_system_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.TestAccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudSystemGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudSystemGroupConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudSystemGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test-acc-system-group"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test system group for acceptance tests"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				Config: testAccJumpCloudSystemGroupConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudSystemGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test-acc-system-group-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test system group"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
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

func testAccJumpCloudSystemGroupConfig() string {
	return `
resource "jumpcloud_system_group" "test" {
  name        = "test-acc-system-group"
  description = "Test system group for acceptance tests"
}
`
}

func testAccJumpCloudSystemGroupConfigUpdated() string {
	return `
resource "jumpcloud_system_group" "test" {
  name        = "test-acc-system-group-updated"
  description = "Updated test system group"
}
`
}

func testAccCheckJumpCloudSystemGroupExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		// Add implementation to check if the system group actually exists in JumpCloud
		// This will depend on how you've structured your test setup

		return nil
	}
}

func testAccCheckJumpCloudSystemGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_system_group" {
			continue
		}

		// Add implementation to check if the system group has been destroyed in JumpCloud
		// This will depend on how you've structured your test setup

		return nil
	}

	return nil
}
