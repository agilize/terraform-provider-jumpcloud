package users

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccJumpCloudUserGroupMembership(t *testing.T) {
	var resourceName = "jumpcloud_user_group_membership.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckJumpCloudUserGroupMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudUserGroupMembershipConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudUserGroupMembershipExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "user_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "user_id"),
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

func testAccJumpCloudUserGroupMembershipConfig() string {
	return `
resource "jumpcloud_user" "test" {
  username   = "test-acc-membership"
  email      = "test-acc-membership@example.com"
  firstname  = "Test"
  lastname   = "User"
}

resource "jumpcloud_user_group" "test" {
  name        = "test-acc-membership-group"
  description = "Test group for acceptance tests"
}

resource "jumpcloud_user_group_membership" "test" {
  user_group_id = jumpcloud_user_group.test.id
  user_id       = jumpcloud_user.test.id
}
`
}

func testAccCheckJumpCloudUserGroupMembershipExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		// Add implementation to check if the membership actually exists in JumpCloud
		// This will depend on how you've structured your test setup

		return nil
	}
}

func testAccCheckJumpCloudUserGroupMembershipDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_user_group_membership" {
			continue
		}

		// Add implementation to check if the membership has been destroyed in JumpCloud
		// This will depend on how you've structured your test setup

		return nil
	}

	return nil
}

// Helper functions
func testAccPreCheck(t *testing.T) {
	// Add any pre-check logic needed for the tests
}

var testAccProviders map[string]*schema.Provider
