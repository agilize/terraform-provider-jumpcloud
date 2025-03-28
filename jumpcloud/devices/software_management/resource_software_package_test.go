package software_management

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// providerFactories is a map of provider factory functions for testing
var packageProviderFactories = map[string]func() (*schema.Provider, error){
	"jumpcloud": func() (*schema.Provider, error) {
		return jctest.TestAccProviders["jumpcloud"], nil
	},
}

// TestResourceSoftwarePackageSchema tests the schema structure of the software package resource
func TestResourceSoftwarePackageSchema(t *testing.T) {
	t.Skip("Skipping schema test until all resources are implemented")
}

// Additional acceptance tests would be added here
// For example:
/*
func TestAccResourceSoftwarePackage_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftwarePackageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSoftwarePackageConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftwarePackageExists("jumpcloud_software_package.test"),
					resource.TestCheckResourceAttr(
						"jumpcloud_software_package.test", "name", "test-package"),
				),
			},
		},
	})
}

func testAccSoftwarePackageConfig_basic() string {
	return `
resource "jumpcloud_software_package" "test" {
  name        = "test-package"
  description = "Test package"
  version     = "1.0.0"
  type        = "windows"
  url         = "https://example.com/package.msi"
}
`
}
*/

func TestAccJumpCloudSoftwarePackage(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.AccPreCheck(t) },
		ProviderFactories: packageProviderFactories,
		// ... existing code ...
	})
}
