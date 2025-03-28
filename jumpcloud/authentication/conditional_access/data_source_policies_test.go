package authentication

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudDataSourceConditionalAccessPolicies_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceConditionalAccessPoliciesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_conditional_access_policies.all", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_conditional_access_policies.all", "policies.#"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceConditionalAccessPoliciesConfig_basic() string {
	return `
data "jumpcloud_conditional_access_policies" "all" {
}
`
}

func TestAccJumpCloudDataSourceConditionalAccessPolicies_filtered(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceConditionalAccessPoliciesConfig_filtered(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_conditional_access_policies.filtered", "id"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceConditionalAccessPoliciesConfig_filtered() string {
	return `
data "jumpcloud_conditional_access_policies" "filtered" {
  filter {
    field    = "status"
    operator = "eq"
    value    = "active"
  }
  
  sort {
    field     = "name"
    direction = "asc"
  }
}

output "active_policy_count" {
  value = length(data.jumpcloud_conditional_access_policies.filtered.policies)
}
`
}

func TestAccJumpCloudDataSourceConditionalAccessPolicies_withPolicy(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceConditionalAccessPoliciesConfig_withPolicy(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_conditional_access_policies.with_policy", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_conditional_access_policies.with_policy", "policies.#"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceConditionalAccessPoliciesConfig_withPolicy() string {
	return `
resource "jumpcloud_conditional_access_policy" "test" {
  name        = "Test Conditional Access Policy"
  description = "Test policy created for data source test"
  status      = "active"
  conditions  = jsonencode({
    device_platforms = ["windows", "macos"]
    locations = ["allowed_locations"]
  })
  actions = jsonencode({
    block_access = true
    require_mfa = true
  })
}

data "jumpcloud_conditional_access_policies" "with_policy" {
  filter {
    field    = "name"
    operator = "eq"
    value    = jumpcloud_conditional_access_policy.test.name
  }
}

output "found_policy_id" {
  value = length(data.jumpcloud_conditional_access_policies.with_policy.policies) > 0 ? data.jumpcloud_conditional_access_policies.with_policy.policies[0].id : "not_found"
}
`
}
