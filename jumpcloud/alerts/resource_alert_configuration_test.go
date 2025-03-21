package alerts

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/testing"
)

// TestResourceAlertConfigurationSchema tests the schema structure of the alert configuration resource
func TestResourceAlertConfigurationSchema(t *testing.T) {
	s := ResourceAlertConfiguration()

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

	if s.Schema["type"] == nil {
		t.Error("Expected type in schema, but it does not exist")
	}
	if s.Schema["type"].Type != schema.TypeString {
		t.Error("Expected type to be of type string")
	}
	if !s.Schema["type"].Required {
		t.Error("Expected type to be required")
	}

	if s.Schema["conditions"] == nil {
		t.Error("Expected conditions in schema, but it does not exist")
	}
	if s.Schema["conditions"].Type != schema.TypeString {
		t.Error("Expected conditions to be of type string")
	}
	if !s.Schema["conditions"].Required {
		t.Error("Expected conditions to be required")
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

	if s.Schema["enabled"] == nil {
		t.Error("Expected enabled in schema, but it does not exist")
	}
	if s.Schema["enabled"].Type != schema.TypeBool {
		t.Error("Expected enabled to be of type bool")
	}
	if s.Schema["enabled"].Required {
		t.Error("Expected enabled to be optional")
	}

	// Test computed fields
	if s.Schema["created"] == nil {
		t.Error("Expected created in schema, but it does not exist")
	}
	if s.Schema["created"].Type != schema.TypeString {
		t.Error("Expected created to be of type string")
	}
	if !s.Schema["created"].Computed {
		t.Error("Expected created to be computed")
	}

	if s.Schema["updated"] == nil {
		t.Error("Expected updated in schema, but it does not exist")
	}
	if s.Schema["updated"].Type != schema.TypeString {
		t.Error("Expected updated to be of type string")
	}
	if !s.Schema["updated"].Computed {
		t.Error("Expected updated to be computed")
	}
}

// TestIsNotFoundError tests the isNotFoundError helper function
func TestIsNotFoundError(t *testing.T) {
	// Test error that matches
	err := fmt.Errorf("status code 404")
	if !isNotFoundError(err) {
		t.Error("Expected error to be detected as 'not found', but it wasn't")
	}

	// Test error that doesn't match
	otherErr := fmt.Errorf("status code 500")
	if isNotFoundError(otherErr) {
		t.Error("Expected error not to be detected as 'not found', but it was")
	}

	// Test nil error
	if isNotFoundError(nil) {
		t.Error("Expected nil error not to be detected as 'not found', but it was")
	}
}

// Acceptance testing
func TestAccResourceAlertConfiguration_basic(t *testing.T) {
	resourceName := "jumpcloud_alert_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { jctest.TestAccPreCheck(t) },
		Providers:    jctest.TestAccProviders,
		CheckDestroy: testAccCheckJumpCloudAlertConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudAlertConfigurationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudAlertConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test-alert-config"),
					resource.TestCheckResourceAttr(resourceName, "type", "system_metric"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
				),
			},
		},
	})
}

func TestAccResourceAlertConfiguration_update(t *testing.T) {
	resourceName := "jumpcloud_alert_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { jctest.TestAccPreCheck(t) },
		Providers:    jctest.TestAccProviders,
		CheckDestroy: testAccCheckJumpCloudAlertConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudAlertConfigurationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudAlertConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test-alert-config"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test alert configuration"),
				),
			},
			{
				Config: testAccJumpCloudAlertConfigurationConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudAlertConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "updated-test-alert-config"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test alert configuration"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudAlertConfigurationDestroy(s *terraform.State) error {
	// Implementation would verify that the resource is deleted on the API side
	// This is a placeholder for the actual implementation
	return nil
}

func testAccCheckJumpCloudAlertConfigurationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Alert Configuration ID is set")
		}

		return nil
	}
}

func testAccJumpCloudAlertConfigurationConfig_basic() string {
	return `
resource "jumpcloud_alert_configuration" "test" {
  name        = "test-alert-config"
  description = "Test alert configuration"
  type        = "system_metric"
  enabled     = true
  conditions  = jsonencode({
    cpu_usage = {
      threshold = 90
      operator  = ">"
      duration  = "5m"
    }
  })
  severity = "medium"
}
`
}

func testAccJumpCloudAlertConfigurationConfig_update() string {
	return `
resource "jumpcloud_alert_configuration" "test" {
  name        = "updated-test-alert-config"
  description = "Updated test alert configuration"
  type        = "system_metric"
  enabled     = true
  conditions  = jsonencode({
    cpu_usage = {
      threshold = 95
      operator  = ">"
      duration  = "10m"
    }
  })
  severity = "high"
}
`
}
