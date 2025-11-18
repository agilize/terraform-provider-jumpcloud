package mfa_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud"
)

func TestAccJumpCloudMFASettings_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckJumpCloudMFASettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudMFASettingsConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMFASettingsExists("jumpcloud_mfa_settings.test"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_settings.test", "system_insights_enrolled", "true"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_settings.test", "exclusion_window_days", "7"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_settings.test", "enabled_methods.#", "2"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_settings.test", "enabled_methods.0", "totp"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_settings.test", "enabled_methods.1", "push"),
				),
			},
			{
				ResourceName:      "jumpcloud_mfa_settings.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccJumpCloudMFASettings_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckJumpCloudMFASettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudMFASettingsConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMFASettingsExists("jumpcloud_mfa_settings.test"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_settings.test", "system_insights_enrolled", "true"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_settings.test", "exclusion_window_days", "7"),
				),
			},
			{
				Config: testAccJumpCloudMFASettingsConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMFASettingsExists("jumpcloud_mfa_settings.test"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_settings.test", "system_insights_enrolled", "false"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_settings.test", "exclusion_window_days", "14"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_settings.test", "enabled_methods.#", "3"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_settings.test", "enabled_methods.0", "totp"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_settings.test", "enabled_methods.1", "push"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_settings.test", "enabled_methods.2", "sms"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudMFASettingsExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MFA Settings ID is set")
		}

		// Add logic to retrieve the MFA settings and validate it exists
		// This would typically involve making an API call to JumpCloud

		return nil
	}
}

func testAccCheckJumpCloudMFASettingsDestroy(s *terraform.State) error {
	// Add logic to verify that the MFA settings have been reset to default values
	// This would typically involve making an API call to JumpCloud

	return nil
}

func testAccJumpCloudMFASettingsConfig_basic() string {
	return `
resource "jumpcloud_mfa_settings" "test" {
  system_insights_enrolled = true
  exclusion_window_days    = 7
  enabled_methods          = ["totp", "push"]
}
`
}

func testAccJumpCloudMFASettingsConfig_update() string {
	return `
resource "jumpcloud_mfa_settings" "test" {
  system_insights_enrolled = false
  exclusion_window_days    = 14
  enabled_methods          = ["totp", "push", "sms"]
}
`
}

// Helper functions for testing
var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = jumpcloud.Provider()
	testAccProviders = map[string]*schema.Provider{
		"jumpcloud": testAccProvider,
	}
}

func testAccPreCheck(t *testing.T) {
	// Add any pre-check validation required before running tests
}
