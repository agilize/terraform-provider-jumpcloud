package devices

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// providerFactories is a map of provider factory functions for testing
var dataSourceProviderFactories = map[string]func() (*schema.Provider, error){
	"jumpcloud": func() (*schema.Provider, error) {
		return jctest.TestAccProviders["jumpcloud"], nil
	},
}

func TestAccDataSourceSystem_basic(t *testing.T) {
	resourceName := "data.jumpcloud_system.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: dataSourceProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSystemConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceSystemExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "display_name"),
				),
			},
		},
	})
}

func testAccCheckDataSourceSystemExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		return nil
	}
}

func testAccDataSourceSystemConfig() string {
	return `
	resource "jumpcloud_system" "test" {
		display_name = "test-system"
		description = "Created for acceptance testing"
		allow_ssh_root_login = false
		allow_ssh_password_authentication = true
		allow_multi_factor_authentication = true
		tags = ["terraform", "test"]
	}

	data "jumpcloud_system" "test" {
		system_id = jumpcloud_system.test.id
	}
	`
}

// Add additional test functions for searching by display name, attribute filtering, etc.
