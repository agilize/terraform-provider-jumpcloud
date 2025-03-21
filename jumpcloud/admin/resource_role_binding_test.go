package admin

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/testing"
)

func TestAccResourceAdminRoleBinding_basic(t *testing.T) {
	rEmail := fmt.Sprintf("terraform-test-%s@example.com", acctest.RandString(8))
	rName := fmt.Sprintf("terraform-test-%s", acctest.RandString(8))
	resourceName := "jumpcloud_admin_role_binding.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAdminRoleBindingConfig_basic(rEmail, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "admin_user_id"),
					resource.TestCheckResourceAttrSet(resourceName, "role_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created"),
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

func testAccResourceAdminRoleBindingConfig_basic(email, roleName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_admin_user" "test" {
  email         = "%s"
  first_name    = "Test"
  last_name     = "Admin"
  status        = "active"
  is_mfa_enabled = true
}

resource "jumpcloud_admin_role" "test" {
  name        = "%s"
  description = "Test role for binding"
  type        = "custom"
  scope       = "org"
  permissions = [
    "read:admin_users",
    "read:admin_roles"
  ]
}

resource "jumpcloud_admin_role_binding" "test" {
  admin_user_id = jumpcloud_admin_user.test.id
  role_id       = jumpcloud_admin_role.test.id
}
`, email, roleName)
}
