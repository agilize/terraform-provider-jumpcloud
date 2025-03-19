package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSoftwareUpdatePolicies_basic(t *testing.T) {
	resourceName := "data.jumpcloud_software_update_policies.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSoftwareUpdatePoliciesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "total"),
					resource.TestCheckResourceAttr(resourceName, "limit", "10"),
				),
			},
		},
	})
}

func TestAccDataSourceSoftwareUpdatePolicies_filtered(t *testing.T) {
	resourceName := "data.jumpcloud_software_update_policies.filtered"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSoftwareUpdatePoliciesConfig_filtered(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "total"),
					resource.TestCheckResourceAttr(resourceName, "os_family", "linux"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "limit", "5"),
				),
			},
		},
	})
}

func testAccDataSourceSoftwareUpdatePoliciesConfig_basic() string {
	return `
data "jumpcloud_software_update_policies" "test" {
  limit = 10
}
`
}

func testAccDataSourceSoftwareUpdatePoliciesConfig_filtered() string {
	return `
data "jumpcloud_software_update_policies" "filtered" {
  os_family = "linux"
  enabled   = true
  limit     = 5
  sort      = "name"
  sort_dir  = "asc"
}
`
}
