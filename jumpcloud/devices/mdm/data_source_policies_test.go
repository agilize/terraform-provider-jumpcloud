package mdm_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudDataSourceMDMPolicies_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceMDMPoliciesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_mdm_policies.all", "id"),
					// The following check may fail if there are no MDM policies, but is useful
					// if at least one policy exists in the test environment
					resource.TestCheckResourceAttrSet("data.jumpcloud_mdm_policies.all", "policies.#"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceMDMPoliciesConfig_basic() string {
	return `
data "jumpcloud_mdm_policies" "all" {
}
`
}

func TestAccJumpCloudDataSourceMDMPolicies_filtered(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceMDMPoliciesConfig_filtered(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_mdm_policies.ios", "id"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceMDMPoliciesConfig_filtered() string {
	return `
data "jumpcloud_mdm_policies" "ios" {
  filter {
    field    = "platform"
    operator = "eq"
    value    = "ios"
  }
  
  sort {
    field     = "name"
    direction = "asc"
  }
}

output "ios_policy_count" {
  value = length(data.jumpcloud_mdm_policies.ios.policies)
}
`
}

func TestAccJumpCloudDataSourceMDMPolicies_withPolicy(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceMDMPoliciesConfig_withPolicy(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_mdm_policies.with_policy", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_mdm_policies.with_policy", "policies.#"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceMDMPoliciesConfig_withPolicy() string {
	return `
resource "jumpcloud_mdm_policy" "test" {
  name        = "Test MDM Policy for Data Source"
  description = "Test policy created for data source test"
  platform    = "ios"
  settings    = jsonencode({
    passcode_required: true,
    passcode_min_length: 8
  })
  scope_type = "all"
}

data "jumpcloud_mdm_policies" "with_policy" {
  filter {
    field    = "name"
    operator = "eq"
    value    = jumpcloud_mdm_policy.test.name
  }
}

output "found_policy_id" {
  value = length(data.jumpcloud_mdm_policies.with_policy.policies) > 0 ? data.jumpcloud_mdm_policies.with_policy.policies[0].id : "not_found"
}
`
}
