package admin_users

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/testing"
)

func TestAccDataSourceAdminUsers_basic(t *testing.T) {
	rEmail := fmt.Sprintf("terraform-test-%s@example.com", acctest.RandString(8))
	dataSourceName := "data.jumpcloud_admin_users.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAdminUsersConfig_basic(rEmail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "users.#"),
					// We can add more specific checks if desired, but since this is
					// a data source that may return multiple users, we're just verifying
					// that the data source itself works.
				),
			},
		},
	})
}

func TestAccDataSourceAdminUsers_filtered(t *testing.T) {
	rEmail := fmt.Sprintf("terraform-test-%s@example.com", acctest.RandString(8))
	dataSourceName := "data.jumpcloud_admin_users.filtered"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAdminUsersConfig_filtered(rEmail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "users.#"),
					// Add specific checks for the filtered data source
				),
			},
		},
	})
}

func testAccDataSourceAdminUsersConfig_basic(email string) string {
	return fmt.Sprintf(`
resource "jumpcloud_admin_user" "test" {
  email         = "%s"
  first_name    = "Test"
  last_name     = "Admin"
  status        = "active"
  is_mfa_enabled = true
}

data "jumpcloud_admin_users" "test" {}
`, email)
}

func testAccDataSourceAdminUsersConfig_filtered(email string) string {
	return fmt.Sprintf(`
resource "jumpcloud_admin_user" "test" {
  email         = "%s"
  first_name    = "Test"
  last_name     = "Admin"
  status        = "active"
  is_mfa_enabled = true
}

# Wait for the user to be created before attempting to filter by it
resource "time_sleep" "wait_30_seconds" {
  depends_on = [jumpcloud_admin_user.test]
  create_duration = "30s"
}

data "jumpcloud_admin_users" "filtered" {
  depends_on = [time_sleep.wait_30_seconds]
  
  filter {
    status = "active"
    search = "Test Admin"
  }
}
`, email)
}
