package admin_roles

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/testing"
)

func TestAccResourceAdminRole_basic(t *testing.T) {
	rName := fmt.Sprintf("terraform-test-%s", acctest.RandString(8))
	resourceName := "jumpcloud_admin_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAdminRoleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Test role for Terraform"),
					resource.TestCheckResourceAttr(resourceName, "type", "custom"),
					resource.TestCheckResourceAttr(resourceName, "scope", "org"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceAdminRole_update(t *testing.T) {
	rName := fmt.Sprintf("terraform-test-%s", acctest.RandString(8))
	resourceName := "jumpcloud_admin_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAdminRoleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Test role for Terraform"),
				),
			},
			{
				Config: testAccResourceAdminRoleConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test role for Terraform"),
					resource.TestCheckResourceAttr(resourceName, "scope", "org"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "3"),
				),
			},
		},
	})
}

func testAccResourceAdminRoleConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_admin_role" "test" {
  name        = "%s"
  description = "Test role for Terraform"
  type        = "custom"
  scope       = "org"
  permissions = [
    "read:admin_users",
    "read:admin_roles"
  ]
}
`, name)
}

func testAccResourceAdminRoleConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_admin_role" "test" {
  name        = "%s"
  description = "Updated test role for Terraform"
  type        = "custom"
  scope       = "org"
  permissions = [
    "read:admin_users",
    "read:admin_roles",
    "list:admin_audit_logs"
  ]
}
`, name)
}
