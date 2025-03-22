package systems

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/testing"
)

// TestResourceSystemSchema tests the schema structure of the system resource
func TestResourceSystemSchema(t *testing.T) {
	s := ResourceSystem()

	// Test required fields
	if s.Schema["display_name"] == nil {
		t.Error("Expected display_name in schema, but it does not exist")
	}
	if s.Schema["display_name"].Type != schema.TypeString {
		t.Error("Expected display_name to be of type string")
	}
	if !s.Schema["display_name"].Required {
		t.Error("Expected display_name to be required")
	}

	// Test optional fields
	if s.Schema["allow_ssh_root_login"] == nil {
		t.Error("Expected allow_ssh_root_login in schema, but it does not exist")
	}
	if s.Schema["allow_ssh_root_login"].Type != schema.TypeBool {
		t.Error("Expected allow_ssh_root_login to be of type bool")
	}
	if s.Schema["allow_ssh_root_login"].Required {
		t.Error("Expected allow_ssh_root_login to be optional")
	}

	if s.Schema["allow_ssh_password_authentication"] == nil {
		t.Error("Expected allow_ssh_password_authentication in schema, but it does not exist")
	}
	if s.Schema["allow_ssh_password_authentication"].Type != schema.TypeBool {
		t.Error("Expected allow_ssh_password_authentication to be of type bool")
	}
	if s.Schema["allow_ssh_password_authentication"].Required {
		t.Error("Expected allow_ssh_password_authentication to be optional")
	}

	if s.Schema["allow_multi_factor_authentication"] == nil {
		t.Error("Expected allow_multi_factor_authentication in schema, but it does not exist")
	}
	if s.Schema["allow_multi_factor_authentication"].Type != schema.TypeBool {
		t.Error("Expected allow_multi_factor_authentication to be of type bool")
	}
	if s.Schema["allow_multi_factor_authentication"].Required {
		t.Error("Expected allow_multi_factor_authentication to be optional")
	}

	if s.Schema["tags"] == nil {
		t.Error("Expected tags in schema, but it does not exist")
	}
	if s.Schema["tags"].Type != schema.TypeList {
		t.Error("Expected tags to be of type list")
	}
	if s.Schema["tags"].Required {
		t.Error("Expected tags to be optional")
	}

	if s.Schema["description"] == nil {
		t.Error("Expected description in schema, but it does not exist")
	}
	if s.Schema["description"].Type != schema.TypeString {
		t.Error("Expected description to be of type string")
	}
	if s.Schema["description"].Required {
		t.Error("Expected description to be optional")
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

	if s.Schema["os"] == nil {
		t.Error("Expected os in schema, but it does not exist")
	}
	if s.Schema["os"].Type != schema.TypeString {
		t.Error("Expected os to be of type string")
	}
	if !s.Schema["os"].Computed {
		t.Error("Expected os to be computed")
	}

	if s.Schema["system_type"] == nil {
		t.Error("Expected system_type in schema, but it does not exist")
	}
	if s.Schema["system_type"].Type != schema.TypeString {
		t.Error("Expected system_type to be of type string")
	}
	if !s.Schema["system_type"].Computed {
		t.Error("Expected system_type to be computed")
	}
}

// Test helper functions
func TestHelperFunctions(t *testing.T) {
	// Test expandStringList
	testList := []interface{}{"tag1", "tag2", "tag3"}
	expanded := expandStringList(testList)

	if len(expanded) != 3 {
		t.Errorf("Expected 3 items, got %d", len(expanded))
	}
	if expanded[0] != "tag1" {
		t.Errorf("Expected first item to be 'tag1', got %s", expanded[0])
	}
	if expanded[1] != "tag2" {
		t.Errorf("Expected second item to be 'tag2', got %s", expanded[1])
	}
	if expanded[2] != "tag3" {
		t.Errorf("Expected third item to be 'tag3', got %s", expanded[2])
	}

	// Test flattenStringList
	testStringList := []string{"item1", "item2", "item3"}
	flattened := flattenStringList(testStringList)

	if len(flattened) != 3 {
		t.Errorf("Expected 3 items, got %d", len(flattened))
	}
	if flattened[0] != "item1" {
		t.Errorf("Expected first item to be 'item1', got %v", flattened[0])
	}
	if flattened[1] != "item2" {
		t.Errorf("Expected second item to be 'item2', got %v", flattened[1])
	}
	if flattened[2] != "item3" {
		t.Errorf("Expected third item to be 'item3', got %v", flattened[2])
	}
}

// Acceptance testing
func TestAccResourceSystem_basic(t *testing.T) {
	resourceName := "jumpcloud_system.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { jctest.TestAccPreCheck(t) },
		Providers:    jctest.TestAccProviders,
		CheckDestroy: testAccCheckJumpCloudSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudSystemConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudSystemExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "display_name", "test-system"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test system"),
				),
			},
		},
	})
}

func TestAccResourceSystem_update(t *testing.T) {
	resourceName := "jumpcloud_system.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { jctest.TestAccPreCheck(t) },
		Providers:    jctest.TestAccProviders,
		CheckDestroy: testAccCheckJumpCloudSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudSystemConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudSystemExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "display_name", "test-system"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test system"),
				),
			},
			{
				Config: testAccJumpCloudSystemConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudSystemExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "display_name", "updated-test-system"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test system"),
					resource.TestCheckResourceAttr(resourceName, "allow_ssh_root_login", "true"),
				),
			},
		},
	})
}

func TestAccResourceSystem_tags(t *testing.T) {
	resourceName := "jumpcloud_system.test_tags"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { jctest.TestAccPreCheck(t) },
		Providers:    jctest.TestAccProviders,
		CheckDestroy: testAccCheckJumpCloudSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudSystemConfig_tags(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudSystemExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "display_name", "test-system-tags"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "dev"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.2", "terraform"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudSystemDestroy(s *terraform.State) error {
	// Implementation would verify that the resource is deleted on the API side
	// This is a placeholder for the actual implementation
	return nil
}

func testAccCheckJumpCloudSystemExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No System ID is set")
		}

		return nil
	}
}

func testAccJumpCloudSystemConfig_basic() string {
	return `
resource "jumpcloud_system" "test" {
  display_name = "test-system"
  description  = "Test system"
}
`
}

func testAccJumpCloudSystemConfig_update() string {
	return `
resource "jumpcloud_system" "test" {
  display_name        = "updated-test-system"
  description         = "Updated test system"
  allow_ssh_root_login = true
}
`
}

func testAccJumpCloudSystemConfig_tags() string {
	return `
resource "jumpcloud_system" "test_tags" {
  display_name = "test-system-tags"
  description  = "Test system with tags"
  
  tags = [
    "dev",
    "test",
    "terraform"
  ]
}
`
}
