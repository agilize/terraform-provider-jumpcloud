package admin

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/testing"
)

func TestAccResourceAdminUser_basic(t *testing.T) {
	rEmail := fmt.Sprintf("terraform-test-%s@example.com", acctest.RandString(8))
	resourceName := "jumpcloud_admin_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAdminUserConfig_basic(rEmail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "email", rEmail),
					resource.TestCheckResourceAttr(resourceName, "first_name", "Test"),
					resource.TestCheckResourceAttr(resourceName, "last_name", "Admin"),
					resource.TestCheckResourceAttr(resourceName, "status", "active"),
					resource.TestCheckResourceAttr(resourceName, "is_mfa_enabled", "true"),
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

func TestAccResourceAdminUser_update(t *testing.T) {
	rEmail := fmt.Sprintf("terraform-test-%s@example.com", acctest.RandString(8))
	resourceName := "jumpcloud_admin_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAdminUserConfig_basic(rEmail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "email", rEmail),
					resource.TestCheckResourceAttr(resourceName, "first_name", "Test"),
					resource.TestCheckResourceAttr(resourceName, "last_name", "Admin"),
				),
			},
			{
				Config: testAccResourceAdminUserConfig_updated(rEmail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "email", rEmail),
					resource.TestCheckResourceAttr(resourceName, "first_name", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "last_name", "AdminUser"),
					resource.TestCheckResourceAttr(resourceName, "status", "active"),
				),
			},
		},
	})
}

func testAccResourceAdminUserConfig_basic(email string) string {
	return fmt.Sprintf(`
resource "jumpcloud_admin_user" "test" {
  email         = "%s"
  first_name    = "Test"
  last_name     = "Admin"
  status        = "active"
  is_mfa_enabled = true
}
`, email)
}

func testAccResourceAdminUserConfig_updated(email string) string {
	return fmt.Sprintf(`
resource "jumpcloud_admin_user" "test" {
  email         = "%s"
  first_name    = "Updated"
  last_name     = "AdminUser"
  status        = "active"
  is_mfa_enabled = true
}
`, email)
}
