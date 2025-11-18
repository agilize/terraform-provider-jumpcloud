package software_management

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// TestResourceSoftwareDeploymentSchema tests the schema structure of the software deployment resource
func TestResourceSoftwareDeploymentSchema(t *testing.T) {
	t.Skip("Skipping schema test until all resources are implemented")
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
resource "jumpcloud_software_package" "test" {
  name        = "test-package"
  version     = "1.0.0"
  type        = "windows"
}

resource "jumpcloud_software_deployment" "test" {
  name        = "test-deployment"
  package_id  = jumpcloud_software_package.test.id
  target_type = "system"
  schedule    = {
    type = "immediate"
  }
}
`
}
*/

func TestAccJumpCloudSoftwareDeployment(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		// ... existing code ...
	})
}
