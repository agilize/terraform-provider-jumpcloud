package software_management

import (
	"testing"
)

// TestResourceSoftwareDeploymentSchema tests the schema structure of the software deployment resource
func TestResourceSoftwareDeploymentSchema(t *testing.T) {
	s := ResourceSoftwareDeployment().Schema

	// Required fields
	for _, required := range []string{"name", "package_id", "target_type", "target_ids"} {
		if !s[required].Required {
			t.Errorf("Expected %s to be required", required)
		}
	}

	// Computed fields
	for _, computed := range []string{"id", "status", "progress", "start_time", "end_time", "created", "updated"} {
		if !s[computed].Computed {
			t.Errorf("Expected %s to be computed", computed)
		}
	}

	// Optional fields
	for _, optional := range []string{"description", "schedule", "parameters", "org_id"} {
		if !s[optional].Optional {
			t.Errorf("Expected %s to be optional", optional)
		}
	}

	// Check that target_type has validation for allowed values
	if s["target_type"].ValidateFunc == nil {
		t.Errorf("Expected target_type to have validation for allowed values")
	}
}

// Additional acceptance tests would be added here
// For example:
/*
func TestAccResourceSoftwareDeployment_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftwareDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSoftwareDeploymentConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftwareDeploymentExists("jumpcloud_software_deployment.test"),
					resource.TestCheckResourceAttr(
						"jumpcloud_software_deployment.test", "name", "test-deployment"),
				),
			},
		},
	})
}

func testAccSoftwareDeploymentConfig_basic() string {
	return `
resource "jumpcloud_software_package" "test_pkg" {
  name    = "test-package"
  version = "1.0.0"
  type    = "windows"
}

resource "jumpcloud_system_group" "test_group" {
  name = "test-group"
}

resource "jumpcloud_software_deployment" "test" {
  name        = "test-deployment"
  description = "Test deployment"
  package_id  = jumpcloud_software_package.test_pkg.id
  target_type = "system_group"
  target_ids  = [jumpcloud_system_group.test_group.id]
  schedule    = {
    type = "immediate"
  }
}
`
}
*/
