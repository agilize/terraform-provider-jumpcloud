package mfa_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudDataSourceMFAStats_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceMFAStatsConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_mfa_stats.test", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_mfa_stats.test", "total_users"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_mfa_stats.test", "mfa_enabled_users"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_mfa_stats.test", "users_with_mfa"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_mfa_stats.test", "mfa_enrollment_rate"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_mfa_stats.test", "method_stats.#"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceMFAStatsConfig_basic() string {
	return `
data "jumpcloud_mfa_stats" "test" {}
`
}
