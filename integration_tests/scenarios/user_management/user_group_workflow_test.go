package user_management

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// TestUserGroupWorkflow tests the complete workflow of:
// 1. Creating users
// 2. Creating a user group
// 3. Adding users to the group
// 4. Assigning the group to an application
func TestUserGroupWorkflow(t *testing.T) {
	// Skip integration tests until CI environment is set up
	t.Skip("Skipping integration test until CI environment is set up")

	// Generate random names to avoid conflicts
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	groupName := fmt.Sprintf("test-group-%s", rName)
	user1Name := fmt.Sprintf("test-user1-%s", rName)
	user2Name := fmt.Sprintf("test-user2-%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			commonTesting.AccPreCheck(t)
		},
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testUserGroupWorkflowConfig(groupName, user1Name, user2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					// Check users were created
					resource.TestCheckResourceAttr(
						"jumpcloud_user.user1", "username", user1Name,
					),
					resource.TestCheckResourceAttr(
						"jumpcloud_user.user2", "username", user2Name,
					),

					// Check group was created
					resource.TestCheckResourceAttr(
						"jumpcloud_user_group.test_group", "name", groupName,
					),

					// Check memberships were created
					resource.TestCheckResourceAttrSet(
						"jumpcloud_user_group_membership.user1_membership", "user_id",
					),
					resource.TestCheckResourceAttrSet(
						"jumpcloud_user_group_membership.user2_membership", "user_id",
					),

					// Check application assignment works
					resource.TestCheckResourceAttrSet(
						"jumpcloud_app_catalog_assignment.test_assignment", "application_id",
					),
				),
			},
		},
	})
}

func testUserGroupWorkflowConfig(groupName, user1Name, user2Name, rName string) string {
	return fmt.Sprintf(`
# Create test users
resource "jumpcloud_user" "user1" {
  username   = %q
  email      = "%s@example.com"
  firstname  = "Test"
  lastname   = "User1"
  password   = "Password123!"
}

resource "jumpcloud_user" "user2" {
  username   = %q
  email      = "%s@example.com"
  firstname  = "Test"
  lastname   = "User2"
  password   = "Password123!"
}

# Create a user group
resource "jumpcloud_user_group" "test_group" {
  name        = %q
  description = "Test group for integration tests"
}

# Add users to the group
resource "jumpcloud_user_group_membership" "user1_membership" {
  user_group_id = jumpcloud_user_group.test_group.id
  user_id       = jumpcloud_user.user1.id
}

resource "jumpcloud_user_group_membership" "user2_membership" {
  user_group_id = jumpcloud_user_group.test_group.id
  user_id       = jumpcloud_user.user2.id
}

# Create an app to assign
resource "jumpcloud_app_catalog_application" "test_app" {
  name        = "Integration Test App %s"
  description = "Test application for integration testing"
}

# Assign the group to the application
resource "jumpcloud_app_catalog_assignment" "test_assignment" {
  application_id = jumpcloud_app_catalog_application.test_app.id
  group_id       = jumpcloud_user_group.test_group.id
}
`, user1Name, user1Name, user2Name, user2Name, groupName, rName)
}
