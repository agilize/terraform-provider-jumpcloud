package radius

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

func TestAccDataSourceRadiusServer_basic(t *testing.T) {
	resourceName := "data.jumpcloud_radius_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: dataSourceProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRadiusServerConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceRadiusServerExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "network_source_ip"),
				),
			},
		},
	})
}

func testAccCheckDataSourceRadiusServerExists(resource string) resource.TestCheckFunc {
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

func testAccDataSourceRadiusServerConfig() string {
	return `
	resource "jumpcloud_radius_server" "test" {
		name = "test-radius-server"
		shared_secret = "test-shared-secret"
		network_source_ip = "10.0.0.1"
		mfa_required = false
		user_password_expiration_action = "allow"
		user_lockout_action = "allow"
		user_attribute = "username"
	}

	data "jumpcloud_radius_server" "test" {
		name = jumpcloud_radius_server.test.name
	}
	`
}

// Additional test for searching by ID
func TestAccDataSourceRadiusServer_byID(t *testing.T) {
	resourceName := "data.jumpcloud_radius_server.test_by_id"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: dataSourceProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRadiusServerIDConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceRadiusServerExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "network_source_ip"),
				),
			},
		},
	})
}

func testAccDataSourceRadiusServerIDConfig() string {
	return `
	resource "jumpcloud_radius_server" "test_for_id" {
		name = "test-radius-server-by-id"
		shared_secret = "test-shared-secret-id"
		network_source_ip = "10.0.0.2"
		mfa_required = true
		user_password_expiration_action = "deny"
		user_lockout_action = "deny"
		user_attribute = "email"
	}

	data "jumpcloud_radius_server" "test_by_id" {
		id = jumpcloud_radius_server.test_for_id.id
	}
	`
}
