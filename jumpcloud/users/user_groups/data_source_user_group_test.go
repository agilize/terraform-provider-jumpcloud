package users_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud"
)

// Test helpers
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("JUMPCLOUD_API_KEY"); v == "" {
		t.Skip("JUMPCLOUD_API_KEY must be set for acceptance tests")
	}
}

var testAccProviders map[string]*schema.Provider

func init() {
	testAccProvider := jumpcloud.Provider()
	testAccProviders = map[string]*schema.Provider{
		"jumpcloud": testAccProvider,
	}
}

// Basic test for user group data source
func TestAccDataSourceUserGroup_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC=1 is set")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUserGroupConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_user_group.test", "id"),
					resource.TestCheckResourceAttr("data.jumpcloud_user_group.test", "name", "test-group"),
					resource.TestCheckResourceAttr("data.jumpcloud_user_group.test", "description", "Test user group"),
				),
			},
		},
	})
}

func testAccDataSourceUserGroupConfig_basic() string {
	return `
resource "jumpcloud_user_group" "test" {
  name        = "test-group"
  description = "Test user group"
}

data "jumpcloud_user_group" "test" {
  group_id = jumpcloud_user_group.test.id
  depends_on = [jumpcloud_user_group.test]
}
`
}

// Test lookup by name
func TestAccDataSourceUserGroup_byName(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC=1 is set")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUserGroupConfig_byName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_user_group.test_name", "id"),
					resource.TestCheckResourceAttr("data.jumpcloud_user_group.test_name", "name", "test-group-by-name"),
				),
			},
		},
	})
}

func testAccDataSourceUserGroupConfig_byName() string {
	return `
resource "jumpcloud_user_group" "test_name" {
  name        = "test-group-by-name"
  description = "Test user group by name"
}

data "jumpcloud_user_group" "test_name" {
  name = jumpcloud_user_group.test_name.name
  depends_on = [jumpcloud_user_group.test_name]
}
`
}

// Test with attributes
func TestAccDataSourceUserGroup_withAttributes(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC=1 is set")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUserGroupConfig_withAttributes(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_user_group.test_attrs", "id"),
					resource.TestCheckResourceAttr("data.jumpcloud_user_group.test_attrs", "attributes.department", "IT"),
					resource.TestCheckResourceAttr("data.jumpcloud_user_group.test_attrs", "attributes.location", "Remote"),
				),
			},
		},
	})
}

func testAccDataSourceUserGroupConfig_withAttributes() string {
	return `
resource "jumpcloud_user_group" "test_attrs" {
  name        = "test-group-with-attrs"
  description = "Test user group with attributes"

  attributes = {
    department = "IT"
    location   = "Remote"
  }
}

data "jumpcloud_user_group" "test_attrs" {
  group_id = jumpcloud_user_group.test_attrs.id
  depends_on = [jumpcloud_user_group.test_attrs]
}
`
}
