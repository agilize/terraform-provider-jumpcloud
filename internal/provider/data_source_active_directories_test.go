package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceActiveDirectories_basic(t *testing.T) {
	dataSourceName := "data.jumpcloud_active_directories.all"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceActiveDirectoriesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "directories.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "total"),
				),
			},
		},
	})
}

func TestAccDataSourceActiveDirectories_filtered(t *testing.T) {
	dataSourceName := "data.jumpcloud_active_directories.filtered"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceActiveDirectoriesConfig_filtered(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "directories.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "total"),
				),
			},
		},
	})
}

func testAccDataSourceActiveDirectoriesConfig_basic() string {
	return `
data "jumpcloud_active_directories" "all" {}
`
}

func testAccDataSourceActiveDirectoriesConfig_filtered() string {
	return `
data "jumpcloud_active_directories" "filtered" {
  search  = "example"
  sort    = "name"
  enabled = true
  limit   = 10
}
`
}
