package admin

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAdminUsers_basic(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	dataSourceName := "data.jumpcloud_admin_users.all"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAdminUsersConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "users.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "total"),
				),
			},
		},
	})
}

func TestAccDataSourceAdminUsers_filtered(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	dataSourceName := "data.jumpcloud_admin_users.filtered"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAdminUsersConfig_filtered(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "users.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "total"),
				),
			},
		},
	})
}

func testAccDataSourceAdminUsersConfig_basic() string {
	return `
data "jumpcloud_admin_users" "all" {}
`
}

func testAccDataSourceAdminUsersConfig_filtered() string {
	return `
data "jumpcloud_admin_users" "filtered" {
  filter {
    name  = "email"
    value = "admin@example.com"
    operator = "contains"
  }
}
`
}
