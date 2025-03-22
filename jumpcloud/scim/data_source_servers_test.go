package scim

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccJumpCloudScimServersDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudScimServersDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudScimServersExists("data.jumpcloud_scim_servers.test"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_scim_servers.test", "servers.#"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudScimServersExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		count, ok := rs.Primary.Attributes["servers.#"]
		if !ok {
			return fmt.Errorf("SCIM servers count not found")
		}

		if count == "0" {
			return fmt.Errorf("No SCIM servers found")
		}

		return nil
	}
}

func testAccJumpCloudScimServersDataSourceConfig() string {
	return `
data "jumpcloud_scim_servers" "test" {
  limit = 10
}
`
}

func TestAccJumpCloudScimServersDataSource_filtered(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudScimServersDataSourceConfigFiltered(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudScimServersExists("data.jumpcloud_scim_servers.filtered"),
				),
			},
		},
	})
}

func testAccJumpCloudScimServersDataSourceConfigFiltered() string {
	return `
data "jumpcloud_scim_servers" "filtered" {
  type    = "custom"
  enabled = true
  limit   = 5
}
`
}
