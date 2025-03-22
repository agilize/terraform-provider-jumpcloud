package software_management

import (
	"testing"
)

// TestResourceSoftwareUpdatePolicySchema tests the schema structure of the software update policy resource
func TestResourceSoftwareUpdatePolicySchema(t *testing.T) {
	s := ResourceSoftwareUpdatePolicy().Schema

	// Required fields
	for _, required := range []string{"name", "os_family", "schedule"} {
		if !s[required].Required {
			t.Errorf("Expected %s to be required", required)
		}
	}

	// Computed fields
	for _, computed := range []string{"id", "created", "updated"} {
		if !s[computed].Computed {
			t.Errorf("Expected %s to be computed", computed)
		}
	}

	// Optional fields
	for _, optional := range []string{"description", "enabled", "package_ids", "all_packages", "auto_approve", "targets", "org_id"} {
		if !s[optional].Optional {
			t.Errorf("Expected %s to be optional", optional)
		}
	}

	// Check conflict between package_ids and all_packages
	if len(s["package_ids"].ConflictsWith) == 0 || s["package_ids"].ConflictsWith[0] != "all_packages" {
		t.Errorf("Expected package_ids to conflict with all_packages")
	}

	if len(s["all_packages"].ConflictsWith) == 0 || s["all_packages"].ConflictsWith[0] != "package_ids" {
		t.Errorf("Expected all_packages to conflict with package_ids")
	}
}

// Additional acceptance tests would be added here
// For example:
/*
func TestAccResourceSoftwareUpdatePolicy_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftwareUpdatePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSoftwareUpdatePolicyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftwareUpdatePolicyExists("jumpcloud_software_update_policy.test"),
					resource.TestCheckResourceAttr(
						"jumpcloud_software_update_policy.test", "name", "test-update-policy"),
				),
			},
		},
	})
}

func testAccSoftwareUpdatePolicyConfig_basic() string {
	return `
resource "jumpcloud_software_update_policy" "test" {
  name        = "test-update-policy"
  description = "Test update policy"
  os_family   = "windows"
  enabled     = true
  schedule    = {
    type = "immediate"
  }
  all_packages = true
  auto_approve = false
}
`
}
*/
