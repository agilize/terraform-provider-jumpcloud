package password_policies

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourcePasswordPolicies_basic(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	dataSourceName := "data.jumpcloud_password_policies.all"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
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

func TestAccDataSourcePasswordPolicies_filtered(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	dataSourceName := "data.jumpcloud_password_policies.filtered"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
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

func testAccDataSourcePasswordPoliciesConfig_filtered() string {
	return `
data "jumpcloud_password_policies" "filtered" {
  filter {
    name  = "status"
    value = "active"
  }
}
`
}
