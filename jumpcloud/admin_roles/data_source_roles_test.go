package admin_roles

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/testing"
)

func TestAccDataSourceAdminRoles_basic(t *testing.T) {
	rName := fmt.Sprintf("terraform-test-%s", acctest.RandString(8))
	dataSourceName := "data.jumpcloud_admin_roles.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAdminRolesConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "roles.#"),
					// We can add more specific checks if desired, but since this is
					// a data source that returns multiple roles, we're just verifying
					// that the data source itself works.
				),
			},
		},
	})
}

func TestAccDataSourceAdminRoles_filtered(t *testing.T) {
	rName := fmt.Sprintf("terraform-test-%s", acctest.RandString(8))
	dataSourceName := "data.jumpcloud_admin_roles.filtered"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAdminRolesConfig_filtered(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "roles.#"),
					// Add specific checks for the filtered data source
				),
			},
		},
	})
}

func testAccDataSourceAdminRolesConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_admin_role" "test" {
  name        = "%s"
  description = "Test role for data source"
  type        = "custom"
  scope       = "org"
  permissions = [
    "read:admin_users",
    "read:admin_roles"
  ]
}

data "jumpcloud_admin_roles" "test" {}
`, name)
}

func testAccDataSourceAdminRolesConfig_filtered(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_admin_role" "test" {
  name        = "%s"
  description = "Test role for data source"
  type        = "custom"
  scope       = "org"
  permissions = [
    "read:admin_users",
    "read:admin_roles"
  ]
}

# Wait for the role to be created before attempting to filter by it
resource "time_sleep" "wait_30_seconds" {
  depends_on = [jumpcloud_admin_role.test]
  create_duration = "30s"
}

data "jumpcloud_admin_roles" "filtered" {
  depends_on = [time_sleep.wait_30_seconds]
  
  filter {
    type = "custom"
    scope = "org"
    search = "%s"
  }
}
`, name, name)
}
