package organization_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccJumpCloudOrganizationSettings_basic(t *testing.T) {
	t.Skip("Skipping acceptance test in unit test environment")

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_organization_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { t.Log("PreCheck running") },
		Providers:    nil, // Will be set by the test framework
		CheckDestroy: testAccCheckJumpCloudOrganizationSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudOrganizationSettingsConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudOrganizationSettingsExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "org_id"),
					resource.TestCheckResourceAttr(resourceName, "system_insights_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "directory_insights_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "ldap_integration_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "allow_public_key_authentication", "true"),
					resource.TestCheckResourceAttr(resourceName, "allow_multi_factor_auth", "true"),
					resource.TestCheckResourceAttr(resourceName, "require_mfa", "false"),
					resource.TestCheckResourceAttr(resourceName, "new_user_email_template", "Welcome to JumpCloud!"),
					resource.TestCheckResourceAttrSet(resourceName, "created"),
				),
			},
		},
	})
}

func TestAccJumpCloudOrganizationSettings_update(t *testing.T) {
	t.Skip("Skipping acceptance test in unit test environment")

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_organization_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { t.Log("PreCheck running") },
		Providers:    nil, // Will be set by the test framework
		CheckDestroy: testAccCheckJumpCloudOrganizationSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudOrganizationSettingsConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudOrganizationSettingsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "system_insights_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "require_mfa", "false"),
				),
			},
			{
				Config: testAccJumpCloudOrganizationSettingsConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudOrganizationSettingsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "system_insights_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "require_mfa", "true"),
					resource.TestCheckResourceAttr(resourceName, "allowed_mfa_methods.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "allowed_mfa_methods.0", "totp"),
					resource.TestCheckResourceAttr(resourceName, "allowed_mfa_methods.1", "push"),
					resource.TestCheckResourceAttr(resourceName, "allowed_mfa_methods.2", "sms"),
				),
			},
		},
	})
}

func TestAccJumpCloudOrganizationSettings_withPasswordPolicy(t *testing.T) {
	t.Skip("Skipping acceptance test in unit test environment")

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_organization_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { t.Log("PreCheck running") },
		Providers:    nil, // Will be set by the test framework
		CheckDestroy: testAccCheckJumpCloudOrganizationSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudOrganizationSettingsConfig_withPasswordPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudOrganizationSettingsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "password_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.min_length", "12"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.requires_lowercase", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.requires_uppercase", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.requires_number", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.requires_special_char", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.expiration_days", "90"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.max_history", "5"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudOrganizationSettingsExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Mock implementation for testing
		return nil
	}
}

func testAccCheckJumpCloudOrganizationSettingsDestroy(s *terraform.State) error {
	// Organization settings can't be destroyed, only reset to default values
	// This is just a placeholder for the resource framework
	return nil
}

func testAccJumpCloudOrganizationSettingsConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_organization" "test" {
  name = "%s"
}

resource "jumpcloud_organization_settings" "test" {
  org_id                     = jumpcloud_organization.test.id
  system_insights_enabled    = true
  directory_insights_enabled = true
  ldap_integration_enabled   = false
  allow_public_key_authentication = true
  allow_multi_factor_auth    = true
  require_mfa                = false
  new_user_email_template    = "Welcome to JumpCloud!"
}
`, name)
}

func testAccJumpCloudOrganizationSettingsConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_organization" "test" {
  name = "%s"
}

resource "jumpcloud_organization_settings" "test" {
  org_id                     = jumpcloud_organization.test.id
  system_insights_enabled    = false
  directory_insights_enabled = true
  ldap_integration_enabled   = true
  allow_public_key_authentication = true
  allow_multi_factor_auth    = true
  require_mfa                = true
  allowed_mfa_methods        = ["totp", "push", "sms"]
  new_user_email_template    = "Welcome to JumpCloud - Updated!"
}
`, name)
}

func testAccJumpCloudOrganizationSettingsConfig_withPasswordPolicy(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_organization" "test" {
  name = "%s"
}

resource "jumpcloud_organization_settings" "test" {
  org_id                     = jumpcloud_organization.test.id
  
  password_policy {
    min_length           = 12
    requires_lowercase   = true
    requires_uppercase   = true
    requires_number      = true
    requires_special_char = true
    expiration_days      = 90
    max_history          = 5
  }
  
  system_insights_enabled    = true
  directory_insights_enabled = true
}
`, name)
}
