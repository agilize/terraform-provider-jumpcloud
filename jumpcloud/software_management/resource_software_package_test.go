package software_management

import (
	"testing"
)

// TestResourceSoftwarePackageSchema tests the schema structure of the software package resource
func TestResourceSoftwarePackageSchema(t *testing.T) {
	s := ResourceSoftwarePackage().Schema

	// Required fields
	for _, required := range []string{"name", "version", "type"} {
		if !s[required].Required {
			t.Errorf("Expected %s to be required", required)
		}
	}

	// Computed fields
	for _, computed := range []string{"id", "status", "created", "updated"} {
		if !s[computed].Computed {
			t.Errorf("Expected %s to be computed", computed)
		}
	}

	// Optional fields
	for _, optional := range []string{"description", "url", "file_path", "file_size", "sha256", "md5", "metadata", "parameters", "tags", "org_id"} {
		if !s[optional].Optional {
			t.Errorf("Expected %s to be optional", optional)
		}
	}
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
