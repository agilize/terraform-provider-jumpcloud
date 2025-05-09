package mfa_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccJumpCloudMFAConfiguration_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckJumpCloudMFAConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudMFAConfigurationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMFAConfigurationExists("jumpcloud_mfa_configuration.test"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_configuration.test", "enabled", "true"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_configuration.test", "exclusive_enabled", "false"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_configuration.test", "system_mfa_required", "true"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_configuration.test", "user_portal_mfa", "true"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_configuration.test", "admin_console_mfa", "true"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_configuration.test", "totp_enabled", "true"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_configuration.test", "push_enabled", "true"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_configuration.test", "duo_enabled", "false"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_configuration.test", "fido_enabled", "false"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_configuration.test", "default_mfa_type", "totp"),
				),
			},
			{
				ResourceName:      "jumpcloud_mfa_configuration.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Duo sensitive fields won't be returned in read
				ImportStateVerifyIgnore: []string{
					"duo_secret_key",
					"duo_application_key",
					"duo_integration_key",
				},
			},
		},
	})
}

func TestAccJumpCloudMFAConfiguration_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckJumpCloudMFAConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudMFAConfigurationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMFAConfigurationExists("jumpcloud_mfa_configuration.test"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_configuration.test", "enabled", "true"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_configuration.test", "push_enabled", "true"),
				),
			},
			{
				Config: testAccJumpCloudMFAConfigurationConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMFAConfigurationExists("jumpcloud_mfa_configuration.test"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_configuration.test", "enabled", "true"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_configuration.test", "exclusive_enabled", "true"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_configuration.test", "system_mfa_required", "true"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_configuration.test", "push_enabled", "false"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_configuration.test", "fido_enabled", "true"),
					resource.TestCheckResourceAttr("jumpcloud_mfa_configuration.test", "default_mfa_type", "fido"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudMFAConfigurationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MFA Configuration ID is set")
		}

		// Add logic to retrieve the MFA configuration and validate it exists
		// This would typically involve making an API call to JumpCloud

		return nil
	}
}

func testAccCheckJumpCloudMFAConfigurationDestroy(s *terraform.State) error {
	// Add logic to verify that the MFA configuration has been reset to default values
	// This would typically involve making an API call to JumpCloud

	return nil
}

func testAccJumpCloudMFAConfigurationConfig_basic() string {
	return `
resource "jumpcloud_mfa_configuration" "test" {
  enabled                = true
  exclusive_enabled      = false
  system_mfa_required    = true
  user_portal_mfa        = true
  admin_console_mfa      = true
  totp_enabled           = true
  push_enabled           = true
  duo_enabled            = false
  fido_enabled           = false
  default_mfa_type       = "totp"
}
`
}

func testAccJumpCloudMFAConfigurationConfig_update() string {
	return `
resource "jumpcloud_mfa_configuration" "test" {
  enabled                = true
  exclusive_enabled      = true
  system_mfa_required    = true
  user_portal_mfa        = true
  admin_console_mfa      = true
  totp_enabled           = true
  push_enabled           = false
  duo_enabled            = false
  fido_enabled           = true
  default_mfa_type       = "fido"
}
`
}

// nolint:unused
func testAccJumpCloudMFAConfigurationConfig_withDuo() string {
	return `
resource "jumpcloud_mfa_configuration" "test" {
  enabled                = true
  exclusive_enabled      = false
  system_mfa_required    = true
  user_portal_mfa        = true
  admin_console_mfa      = true
  totp_enabled           = true
  push_enabled           = true
  duo_enabled            = true
  fido_enabled           = false
  default_mfa_type       = "duo"
  duo_api_hostname       = "api-12345678.duosecurity.com"
  duo_secret_key         = "secretkeyexample123456789"
  duo_integration_key    = "integrationkeyexample"
  duo_application_key    = "applicationkeyexample"
}
`
}
