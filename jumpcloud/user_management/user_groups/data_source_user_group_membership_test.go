package usergroups_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccDataSourceUsergroupMembership_basic tests basic data source functionality
func TestAccDataSourceUsergroupMembership_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUsergroupMembershipConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_usergroup_membership.test", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_usergroup_membership.test", "user_group_id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_usergroup_membership.test", "user_id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_usergroup_membership.test", "member"),
				),
			},
		},
	})
}

// TestAccDataSourceUsergroupMembership_member tests checking a member user
func TestAccDataSourceUsergroupMembership_member(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUsergroupMembershipConfig_member(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.jumpcloud_usergroup_membership.test", "member", "true"),
				),
			},
		},
	})
}

// TestAccDataSourceUsergroupMembership_notMember tests checking a non-member user
func TestAccDataSourceUsergroupMembership_notMember(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUsergroupMembershipConfig_notMember(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.jumpcloud_usergroup_membership.test", "member", "false"),
				),
			},
		},
	})
}

func testAccDataSourceUsergroupMembershipConfig_basic() string {
	return `
resource "jumpcloud_usergroup" "test" {
  name = "test-group-membership-ds"
}

resource "jumpcloud_user" "test" {
  username = "testuser-membership-ds@example.com"
  email    = "testuser-membership-ds@example.com"
}

data "jumpcloud_usergroup_membership" "test" {
  user_group_id = jumpcloud_usergroup.test.id
  user_id       = jumpcloud_user.test.id
}
`
}

func testAccDataSourceUsergroupMembershipConfig_member() string {
	return `
resource "jumpcloud_usergroup" "test" {
  name = "test-group-membership-member"
}

resource "jumpcloud_user" "test" {
  username = "testuser-membership-member@example.com"
  email    = "testuser-membership-member@example.com"
}

resource "jumpcloud_usergroup_membership" "test" {
  user_group_id = jumpcloud_usergroup.test.id
  user_id       = jumpcloud_user.test.id
}

data "jumpcloud_usergroup_membership" "test" {
  user_group_id = jumpcloud_usergroup.test.id
  user_id       = jumpcloud_user.test.id
  depends_on    = [jumpcloud_usergroup_membership.test]
}
`
}

func testAccDataSourceUsergroupMembershipConfig_notMember() string {
	return `
resource "jumpcloud_usergroup" "test" {
  name = "test-group-membership-notmember"
}

resource "jumpcloud_user" "test" {
  username = "testuser-membership-notmember@example.com"
  email    = "testuser-membership-notmember@example.com"
}

resource "jumpcloud_user" "other" {
  username = "otheruser-membership-notmember@example.com"
  email    = "otheruser-membership-notmember@example.com"
}

resource "jumpcloud_usergroup_membership" "test" {
  user_group_id = jumpcloud_usergroup.test.id
  user_id       = jumpcloud_user.test.id
}

data "jumpcloud_usergroup_membership" "test" {
  user_group_id = jumpcloud_usergroup.test.id
  user_id       = jumpcloud_user.other.id
  depends_on    = [jumpcloud_usergroup_membership.test]
}
`
}
