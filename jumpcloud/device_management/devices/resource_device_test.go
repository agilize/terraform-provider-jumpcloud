package devices

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// TestResourceDeviceSchema tests the schema structure of the device resource
func TestResourceDeviceSchema(t *testing.T) {
	s := ResourceDevice()

	// Test display_name field (optional + computed for import workflow)
	if s.Schema["display_name"] == nil {
		t.Error("Expected display_name in schema, but it does not exist")
	}
	if s.Schema["display_name"].Type != schema.TypeString {
		t.Error("Expected display_name to be of type string")
	}
	if s.Schema["display_name"].Required {
		t.Error("Expected display_name to be optional (not required)")
	}
	if !s.Schema["display_name"].Optional {
		t.Error("Expected display_name to be optional")
	}
	if !s.Schema["display_name"].Computed {
		t.Error("Expected display_name to be computed")
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

	if s.Schema["device_type"] == nil {
		t.Error("Expected device_type in schema, but it does not exist")
	}
	if s.Schema["device_type"].Type != schema.TypeString {
		t.Error("Expected device_type to be of type string")
	}
	if !s.Schema["device_type"].Computed {
		t.Error("Expected device_type to be computed")
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

	// Test that Importer is defined (required for import workflow)
	if s.Importer == nil {
		t.Error("Expected Importer to be defined for import workflow")
	}
}

// Acceptance testing
func TestAccResourceDevice_basic(t *testing.T) {
	resourceName := "jumpcloud_devices.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDeviceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "display_name", "test-device"),
				),
			},
		},
	})
}

func TestAccResourceDevice_update(t *testing.T) {
	resourceName := "jumpcloud_devices.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDeviceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "display_name", "test-device"),
				),
			},
			{
				Config: testAccJumpCloudDeviceConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "display_name", "updated-test-device"),
					resource.TestCheckResourceAttr(resourceName, "allow_ssh_root_login", "true"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudDeviceDestroy(s *terraform.State) error {
	// Implementation would verify that the resource is deleted on the API side
	// This is a placeholder for the actual implementation
	return nil
}

func testAccCheckJumpCloudDeviceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Device ID is set")
		}

		return nil
	}
}

func testAccJumpCloudDeviceConfig_basic() string {
	return `
resource "jumpcloud_devices" "test" {
  display_name = "test-device"
}
`
}

func testAccJumpCloudDeviceConfig_update() string {
	return `
resource "jumpcloud_devices" "test" {
  display_name          = "updated-test-device"
  allow_ssh_root_login  = true
}
`
}
