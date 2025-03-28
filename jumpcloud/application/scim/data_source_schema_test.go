package scim

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// Definindo as provider factories

func TestAccJumpCloudScimSchemaDataSource_basic(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudScimSchemaDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudScimSchemaExists("data.jumpcloud_scim_schema.test"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_scim_schema.test", "name"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_scim_schema.test", "uri"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_scim_schema.test", "attributes.#"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudScimSchemaExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return nil
		}

		if rs.Primary.ID == "" {
			return nil
		}

		return nil
	}
}

func testAccJumpCloudScimSchemaDataSourceConfig() string {
	return `
data "jumpcloud_scim_schema" "test" {
  uri = "urn:ietf:params:scim:schemas:core:2.0:User"
}
`
}
