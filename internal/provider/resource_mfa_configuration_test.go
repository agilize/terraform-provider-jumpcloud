package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMfaConfiguration_basic(t *testing.T) {
	var mfaID string

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_mfa_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMfaConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMfaConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMfaConfigurationExists(resourceName, &mfaID),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "totp"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "totp_settings.0.digits", "6"),
					resource.TestCheckResourceAttr(resourceName, "totp_settings.0.algorithm", "SHA1"),
					resource.TestCheckResourceAttr(resourceName, "totp_settings.0.period", "30"),
					resource.TestMatchResourceAttr(resourceName, "created", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
					resource.TestMatchResourceAttr(resourceName, "updated", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceMfaConfiguration_update(t *testing.T) {
	var mfaID string

	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := rName + "-updated"
	resourceName := "jumpcloud_mfa_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMfaConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMfaConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMfaConfigurationExists(resourceName, &mfaID),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "totp"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccMfaConfigurationConfig_updated(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMfaConfigurationExists(resourceName, &mfaID),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "type", "totp"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "totp_settings.0.algorithm", "SHA256"),
					resource.TestCheckResourceAttr(resourceName, "totp_settings.0.period", "60"),
				),
			},
		},
	})
}

func testAccCheckMfaConfigurationExists(resourceName string, mfaID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		*mfaID = rs.Primary.ID

		return nil
	}
}

func testAccCheckMfaConfigurationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_mfa_configuration" {
			continue
		}

		// Retrieve the client from the test provider
		client := testAccProvider.Meta().(JumpCloudClient)

		// Check that the MFA configuration no longer exists
		url := fmt.Sprintf("/api/v2/mfa/config/%s", rs.Primary.ID)
		_, err := client.DoRequest("GET", url, nil)

		// The request should return an error if the MFA configuration is destroyed
		if err == nil {
			return fmt.Errorf("JumpCloud MFA configuration %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccMfaConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_mfa_configuration" "test" {
  name    = %q
  type    = "totp"
  enabled = true
  
  totp_settings {
    digits    = 6
    algorithm = "SHA1"
    period    = 30
  }
}
`, rName)
}

func testAccMfaConfigurationConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_mfa_configuration" "test" {
  name    = %q
  type    = "totp"
  enabled = false
  
  totp_settings {
    digits    = 6
    algorithm = "SHA256"
    period    = 60
  }
}
`, rName)
}
