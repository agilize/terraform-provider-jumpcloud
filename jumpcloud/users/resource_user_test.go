package users

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/testing"
)

// TestResourceUserSchema tests the schema structure of the user resource
func TestResourceUserSchema(t *testing.T) {
	s := ResourceUser()

	// Test required fields
	if s.Schema["username"] == nil {
		t.Error("Expected username in schema, but it does not exist")
	}
	if s.Schema["username"].Type != schema.TypeString {
		t.Error("Expected username to be of type string")
	}
	if !s.Schema["username"].Required {
		t.Error("Expected username to be required")
	}

	if s.Schema["email"] == nil {
		t.Error("Expected email in schema, but it does not exist")
	}
	if s.Schema["email"].Type != schema.TypeString {
		t.Error("Expected email to be of type string")
	}
	if !s.Schema["email"].Required {
		t.Error("Expected email to be required")
	}

	// Test optional fields
	if s.Schema["firstname"] == nil {
		t.Error("Expected firstname in schema, but it does not exist")
	}
	if s.Schema["firstname"].Type != schema.TypeString {
		t.Error("Expected firstname to be of type string")
	}
	if s.Schema["firstname"].Required {
		t.Error("Expected firstname to be optional")
	}

	if s.Schema["lastname"] == nil {
		t.Error("Expected lastname in schema, but it does not exist")
	}
	if s.Schema["lastname"].Type != schema.TypeString {
		t.Error("Expected lastname to be of type string")
	}
	if s.Schema["lastname"].Required {
		t.Error("Expected lastname to be optional")
	}

	if s.Schema["password"] == nil {
		t.Error("Expected password in schema, but it does not exist")
	}
	if s.Schema["password"].Type != schema.TypeString {
		t.Error("Expected password to be of type string")
	}
	if s.Schema["password"].Required {
		t.Error("Expected password to be optional")
	}
	if !s.Schema["password"].Sensitive {
		t.Error("Expected password to be sensitive")
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

// Test helper functions
func TestHelperFunctions(t *testing.T) {
	// Test expandAttributes
	testAttrs := map[string]interface{}{
		"department": "IT",
		"location":   "Remote",
		"title":      "Developer",
	}

	expanded := expandAttributes(testAttrs)
	if len(expanded) != 3 {
		t.Errorf("Expected 3 attributes, got %d", len(expanded))
	}
	if expanded["department"] != "IT" {
		t.Errorf("Expected department to be 'IT', got %v", expanded["department"])
	}
	if expanded["location"] != "Remote" {
		t.Errorf("Expected location to be 'Remote', got %v", expanded["location"])
	}
	if expanded["title"] != "Developer" {
		t.Errorf("Expected title to be 'Developer', got %v", expanded["title"])
	}

	// Test flattenAttributes
	attrs := map[string]interface{}{
		"department": "Engineering",
		"count":      42,
		"active":     true,
	}

	flattened := flattenAttributes(attrs)
	if len(flattened) != 3 {
		t.Errorf("Expected 3 attributes, got %d", len(flattened))
	}
	if flattened["department"] != "Engineering" {
		t.Errorf("Expected department to be 'Engineering', got %v", flattened["department"])
	}
	if flattened["count"] != "42" {
		t.Errorf("Expected count to be '42', got %v", flattened["count"])
	}
	if flattened["active"] != "true" {
		t.Errorf("Expected active to be 'true', got %v", flattened["active"])
	}
}

// Acceptance testing
func TestAccResourceUser_basic(t *testing.T) {
	resourceName := "jumpcloud_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { jctest.TestAccPreCheck(t) },
		Providers:    jctest.TestAccProviders,
		CheckDestroy: testAccCheckJumpCloudUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudUserConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "username", "testuser"),
					resource.TestCheckResourceAttr(resourceName, "email", "testuser@example.com"),
					resource.TestCheckResourceAttr(resourceName, "firstname", "Test"),
					resource.TestCheckResourceAttr(resourceName, "lastname", "User"),
				),
			},
		},
	})
}

func TestAccResourceUser_update(t *testing.T) {
	resourceName := "jumpcloud_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { jctest.TestAccPreCheck(t) },
		Providers:    jctest.TestAccProviders,
		CheckDestroy: testAccCheckJumpCloudUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudUserConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "username", "testuser"),
					resource.TestCheckResourceAttr(resourceName, "email", "testuser@example.com"),
				),
			},
			{
				Config: testAccJumpCloudUserConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "email", "updated.testuser@example.com"),
					resource.TestCheckResourceAttr(resourceName, "firstname", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "lastname", "User"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test user"),
				),
			},
		},
	})
}

func TestAccResourceUser_attributes(t *testing.T) {
	resourceName := "jumpcloud_user.test_attrs"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { jctest.TestAccPreCheck(t) },
		Providers:    jctest.TestAccProviders,
		CheckDestroy: testAccCheckJumpCloudUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudUserConfig_attributes(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes.department", "Engineering"),
					resource.TestCheckResourceAttr(resourceName, "attributes.location", "Remote"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudUserDestroy(s *terraform.State) error {
	// Implementation would verify that the resource is deleted on the API side
	// This is a placeholder for the actual implementation
	return nil
}

func testAccCheckJumpCloudUserExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No User ID is set")
		}

		return nil
	}
}

func testAccJumpCloudUserConfig_basic() string {
	return `
resource "jumpcloud_user" "test" {
  username   = "testuser"
  email      = "testuser@example.com"
  firstname  = "Test"
  lastname   = "User"
  password   = "P@ssw0rd123!"
  description = "Test user"
}
`
}

func testAccJumpCloudUserConfig_update() string {
	return `
resource "jumpcloud_user" "test" {
  username   = "testuser"
  email      = "updated.testuser@example.com"
  firstname  = "Updated"
  lastname   = "User"
  password   = "P@ssw0rd123!"
  description = "Updated test user"
}
`
}

func testAccJumpCloudUserConfig_attributes() string {
	return `
resource "jumpcloud_user" "test_attrs" {
  username   = "testuser_attrs"
  email      = "testuser_attrs@example.com"
  firstname  = "Test"
  lastname   = "User"
  password   = "P@ssw0rd123!"
  
  attributes = {
    department = "Engineering"
    location   = "Remote"
    title      = "Developer"
  }
}
`
}
