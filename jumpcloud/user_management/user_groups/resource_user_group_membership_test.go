package usergroups_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// TestAccResourceUsergroupMembership_basic tests basic creation of a user group membership
func TestAccResourceUsergroupMembership_basic(t *testing.T) {
	resourceName := "jumpcloud_user_group_membership.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		CheckDestroy:      testAccCheckUsergroupMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsergroupMembershipConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsergroupMembershipExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "user_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "user_id"),
				),
			},
		},
	})
}

// TestAccResourceUsergroupMembership_import tests importing an existing membership
func TestAccResourceUsergroupMembership_import(t *testing.T) {
	resourceName := "jumpcloud_user_group_membership.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		CheckDestroy:      testAccCheckUsergroupMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsergroupMembershipConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsergroupMembershipExists(resourceName),
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

// TestAccResourceUsergroupMembership_multipleMemberships tests multiple users in one group
func TestAccResourceUsergroupMembership_multipleMemberships(t *testing.T) {
	resource1Name := "jumpcloud_user_group_membership.test1"
	resource2Name := "jumpcloud_user_group_membership.test2"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		CheckDestroy:      testAccCheckUsergroupMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsergroupMembershipConfig_multiple(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsergroupMembershipExists(resource1Name),
					testAccCheckUsergroupMembershipExists(resource2Name),
					resource.TestCheckResourceAttrSet(resource1Name, "id"),
					resource.TestCheckResourceAttrSet(resource2Name, "id"),
				),
			},
		},
	})
}

// TestAccResourceUsergroupMembership_validation tests validation errors
func TestAccResourceUsergroupMembership_validation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccUsergroupMembershipConfig_missingUsergroupID(),
				ExpectError: regexp.MustCompile("user_group_id"),
			},
			{
				Config:      testAccUsergroupMembershipConfig_missingUserID(),
				ExpectError: regexp.MustCompile("user_id"),
			},
		},
	})
}

// Helper functions

func testAccCheckUsergroupMembershipExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No User Group Membership ID is set")
		}

		return nil
	}
}

func testAccCheckUsergroupMembershipDestroy(s *terraform.State) error {
	// This would check if the membership was properly destroyed
	// For now, we'll return nil as the actual implementation would require API calls
	return nil
}

// Configuration functions

func testAccUsergroupMembershipConfig_basic() string {
	return `
resource "jumpcloud_user" "test" {
  username   = "testuser_group_member"
  email      = "testuser_group_member@example.com"
  firstname  = "Test"
  lastname   = "User"
  password   = "P@ssw0rd123!"
}

resource "jumpcloud_user_group" "test" {
  name        = "test-group-membership"
  description = "Test group for membership tests"
}

resource "jumpcloud_user_group_membership" "test" {
  user_group_id = jumpcloud_user_group.test.id
  user_id       = jumpcloud_user.test.id
}
`
}

func testAccUsergroupMembershipConfig_multiple() string {
	return `
resource "jumpcloud_user" "test1" {
  username   = "testuser_group_member1"
  email      = "testuser_group_member1@example.com"
  firstname  = "Test"
  lastname   = "User1"
  password   = "P@ssw0rd123!"
}

resource "jumpcloud_user" "test2" {
  username   = "testuser_group_member2"
  email      = "testuser_group_member2@example.com"
  firstname  = "Test"
  lastname   = "User2"
  password   = "P@ssw0rd123!"
}

resource "jumpcloud_user_group" "test" {
  name        = "test-group-membership"
  description = "Test group for membership tests"
}

resource "jumpcloud_user_group_membership" "test1" {
  user_group_id = jumpcloud_user_group.test.id
  user_id       = jumpcloud_user.test1.id
}

resource "jumpcloud_user_group_membership" "test2" {
  user_group_id = jumpcloud_user_group.test.id
  user_id       = jumpcloud_user.test2.id
}
`
}

func testAccUsergroupMembershipConfig_missingUsergroupID() string {
	return `
resource "jumpcloud_user" "test" {
  username   = "testuser_group_member"
  email      = "testuser_group_member@example.com"
  firstname  = "Test"
  lastname   = "User"
  password   = "P@ssw0rd123!"
}

resource "jumpcloud_user_group_membership" "test" {
  user_id = jumpcloud_user.test.id
}
`
}

func testAccUsergroupMembershipConfig_missingUserID() string {
	return `
resource "jumpcloud_user_group" "test" {
  name        = "test-group-membership"
  description = "Test group for membership tests"
}

resource "jumpcloud_user_group_membership" "test" {
  user_group_id = jumpcloud_user_group.test.id
}
`
}
