package system_groups

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

func TestAccDataSourceSystemGroup_basic(t *testing.T) {
	resourceName := "data.jumpcloud_system_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: dataSourceProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSystemGroupConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceSystemGroupExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
				),
			},
		},
	})
}

func testAccCheckDataSourceSystemGroupExists(resource string) resource.TestCheckFunc {
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

func testAccDataSourceSystemGroupConfig() string {
	return `
	resource "jumpcloud_system_group" "test" {
		name = "test-system-group"
		description = "Created for acceptance testing"
		attributes = {
			"environment" = "testing"
		}
	}

	data "jumpcloud_system_group" "test" {
		name = jumpcloud_system_group.test.name
	}
	`
}

// Additional test for searching by ID
func TestAccDataSourceSystemGroup_byID(t *testing.T) {
	resourceName := "data.jumpcloud_system_group.test_by_id"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: dataSourceProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSystemGroupIDConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceSystemGroupExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
				),
			},
		},
	})
}

func testAccDataSourceSystemGroupIDConfig() string {
	return `
	resource "jumpcloud_system_group" "test_for_id" {
		name = "test-system-group-by-id"
		description = "Created for ID-based testing"
	}

	data "jumpcloud_system_group" "test_by_id" {
		id = jumpcloud_system_group.test_for_id.id
	}
	`
}
