package users

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/testing"
)

// TestResourceUserGroupSchema tests the schema structure of the user group resource
func TestResourceUserGroupSchema(t *testing.T) {
	s := ResourceUserGroup()

	// Test required fields
	if s.Schema["name"] == nil {
		t.Error("Expected name in schema, but it does not exist")
	}
	if s.Schema["name"].Type != schema.TypeString {
		t.Error("Expected name to be of type string")
	}
	if !s.Schema["name"].Required {
		t.Error("Expected name to be required")
	}

	// Test optional fields
	if s.Schema["description"] == nil {
		t.Error("Expected description in schema, but it does not exist")
	}
	if s.Schema["description"].Type != schema.TypeString {
		t.Error("Expected description to be of type string")
	}
	if s.Schema["description"].Required {
		t.Error("Expected description to be optional")
	}

	if s.Schema["type"] == nil {
		t.Error("Expected type in schema, but it does not exist")
	}
	if s.Schema["type"].Type != schema.TypeString {
		t.Error("Expected type to be of type string")
	}
	if s.Schema["type"].Required {
		t.Error("Expected type to be optional")
	}

	if s.Schema["attributes"] == nil {
		t.Error("Expected attributes in schema, but it does not exist")
	}
	if s.Schema["attributes"].Type != schema.TypeMap {
		t.Error("Expected attributes to be of type map")
	}
	if s.Schema["attributes"].Required {
		t.Error("Expected attributes to be optional")
	}

	// Test computed fields
	if s.Schema["id"] == nil {
		t.Error("Expected id in schema, but it does not exist")
	}
	if s.Schema["id"].Type != schema.TypeString {
		t.Error("Expected id to be of type string")
	}
	if !s.Schema["id"].Computed {
		t.Error("Expected id to be computed")
	}

	if s.Schema["member_count"] == nil {
		t.Error("Expected member_count in schema, but it does not exist")
	}
	if s.Schema["member_count"].Type != schema.TypeInt {
		t.Error("Expected member_count to be of type int")
	}
	if !s.Schema["member_count"].Computed {
		t.Error("Expected member_count to be computed")
	}

	if s.Schema["created"] == nil {
		t.Error("Expected created in schema, but it does not exist")
	}
	if s.Schema["created"].Type != schema.TypeString {
		t.Error("Expected created to be of type string")
	}
	if !s.Schema["created"].Computed {
		t.Error("Expected created to be computed")
	}
}

// Acceptance testing
func TestAccResourceUserGroup_basic(t *testing.T) {
	resourceName := "jumpcloud_user_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { jctest.TestAccPreCheck(t) },
		Providers:    jctest.TestAccProviders,
		CheckDestroy: testAccCheckJumpCloudUserGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudUserGroupConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudUserGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test-user-group"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test user group"),
				),
			},
		},
	})
}

func TestAccResourceUserGroup_update(t *testing.T) {
	resourceName := "jumpcloud_user_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { jctest.TestAccPreCheck(t) },
		Providers:    jctest.TestAccProviders,
		CheckDestroy: testAccCheckJumpCloudUserGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudUserGroupConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudUserGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test-user-group"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test user group"),
				),
			},
			{
				Config: testAccJumpCloudUserGroupConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudUserGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "updated-test-user-group"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test user group"),
				),
			},
		},
	})
}

func TestAccResourceUserGroup_attributes(t *testing.T) {
	resourceName := "jumpcloud_user_group.test_attrs"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { jctest.TestAccPreCheck(t) },
		Providers:    jctest.TestAccProviders,
		CheckDestroy: testAccCheckJumpCloudUserGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudUserGroupConfig_attributes(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudUserGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes.department", "Engineering"),
					resource.TestCheckResourceAttr(resourceName, "attributes.location", "Remote"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudUserGroupDestroy(s *terraform.State) error {
	// Implementation would verify that the resource is deleted on the API side
	// This is a placeholder for the actual implementation
	return nil
}

func testAccCheckJumpCloudUserGroupExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No User Group ID is set")
		}

		return nil
	}
}

func testAccJumpCloudUserGroupConfig_basic() string {
	return `
resource "jumpcloud_user_group" "test" {
  name        = "test-user-group"
  description = "Test user group"
}
`
}

func testAccJumpCloudUserGroupConfig_update() string {
	return `
resource "jumpcloud_user_group" "test" {
  name        = "updated-test-user-group"
  description = "Updated test user group"
}
`
}

func testAccJumpCloudUserGroupConfig_attributes() string {
	return `
resource "jumpcloud_user_group" "test_attrs" {
  name        = "test-user-group-attrs"
  description = "Test user group with attributes"
  
  attributes = {
    department = "Engineering"
    location   = "Remote"
    category   = "Developers"
  }
}
`
}
