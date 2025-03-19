package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourcePasswordPolicies_basic(t *testing.T) {
	dataSourceName := "data.jumpcloud_password_policies.all"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePasswordPoliciesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "policies.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "total"),
				),
			},
		},
	})
}

func TestAccDataSourcePasswordPolicies_search(t *testing.T) {
	dataSourceName := "data.jumpcloud_password_policies.search"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePasswordPoliciesConfig_search(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "policies.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "total"),
				),
			},
		},
	})
}

func TestAccDataSourcePasswordPolicies_filtered(t *testing.T) {
	dataSourceName := "data.jumpcloud_password_policies.filtered"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePasswordPoliciesConfig_filtered(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "policies.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "total"),
				),
			},
		},
	})
}

func testAccDataSourcePasswordPoliciesConfig_basic() string {
	return `
data "jumpcloud_password_policies" "all" {}
`
}

func testAccDataSourcePasswordPoliciesConfig_search() string {
	return `
data "jumpcloud_password_policies" "search" {
  search = "security"
  limit  = 5
}

output "search_policies" {
  value = data.jumpcloud_password_policies.search.policies
}

output "search_total" {
  value = data.jumpcloud_password_policies.search.total
}
`
}

func testAccDataSourcePasswordPoliciesConfig_filtered() string {
	return `
data "jumpcloud_password_policies" "filtered" {
  name   = "Default"
  status = "active"
  sort   = "name"
  limit  = 10
}
`
}
