package mfa_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccJumpCloudDataSourceMFASettings_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceMFASettingsConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_mfa_settings.test", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_mfa_settings.test", "system_insights_enrolled"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_mfa_settings.test", "exclusion_window_days"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_mfa_settings.test", "enabled_methods.#"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_mfa_settings.test", "updated"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceMFASettingsConfig_basic() string {
	return `
data "jumpcloud_mfa_settings" "test" {
}
`
}

func TestAccJumpCloudDataSourceMFASettings_withResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceMFASettingsConfig_withResource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.jumpcloud_mfa_settings.test", "system_insights_enrolled",
						"jumpcloud_mfa_settings.test", "system_insights_enrolled",
					),
					resource.TestCheckResourceAttrPair(
						"data.jumpcloud_mfa_settings.test", "exclusion_window_days",
						"jumpcloud_mfa_settings.test", "exclusion_window_days",
					),
					resource.TestCheckResourceAttrPair(
						"data.jumpcloud_mfa_settings.test", "enabled_methods.#",
						"jumpcloud_mfa_settings.test", "enabled_methods.#",
					),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceMFASettingsConfig_withResource() string {
	return `
resource "jumpcloud_mfa_settings" "test" {
  system_insights_enrolled = true
  exclusion_window_days    = 7
  enabled_methods          = ["totp", "push"]
}

data "jumpcloud_mfa_settings" "test" {
  depends_on = [jumpcloud_mfa_settings.test]
}
`
}
