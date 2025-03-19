package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMfaStats_basic(t *testing.T) {
	dataSourceName := "data.jumpcloud_mfa_stats.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMfaStatsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "total_users"),
					resource.TestCheckResourceAttrSet(dataSourceName, "mfa_configured_users"),
					resource.TestCheckResourceAttrSet(dataSourceName, "enrollment_percentage"),
					resource.TestCheckResourceAttrSet(dataSourceName, "method_stats.#"),
				),
			},
		},
	})
}

func testAccDataSourceMfaStatsConfig() string {
	return `
data "jumpcloud_mfa_stats" "test" {}
`
}
