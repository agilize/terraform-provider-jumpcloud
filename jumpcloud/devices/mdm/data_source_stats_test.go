package mdm_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudDataSourceMDMStats_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDMStatsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_mdm_stats.test", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_mdm_stats.test", "total_devices"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_mdm_stats.test", "corporate_devices"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_mdm_stats.test", "byod_devices"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_mdm_stats.test", "ios_devices"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_mdm_stats.test", "android_devices"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_mdm_stats.test", "macos_devices"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_mdm_stats.test", "windows_devices"),
				),
			},
		},
	})
}

func testAccDataSourceMDMStatsConfig() string {
	return `
data "jumpcloud_mdm_stats" "test" {
}
`
}
