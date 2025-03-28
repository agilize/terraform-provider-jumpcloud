package password_policies

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccDataSourcePasswordPolicies_basic(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePasswordPoliciesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_password_policies.all", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_password_policies.all", "policies.#"),
				),
			},
		},
	})
}

func TestAccDataSourcePasswordPolicies_filtered(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePasswordPoliciesConfig_filtered(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_password_policies.filtered", "policies.#"),
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
