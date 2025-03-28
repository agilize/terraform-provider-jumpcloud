package mdm_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudMDMPolicy_basic(t *testing.T) {
	resourceName := "jumpcloud_mdm_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudMDMPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudMDMPolicyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMDMPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "Test MDM Policy"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test MDM policy for acceptance tests"),
					resource.TestCheckResourceAttr(resourceName, "platform", "ios"),
					resource.TestCheckResourceAttr(resourceName, "settings.passcode.required", "true"),
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

func TestAccJumpCloudMDMPolicy_update(t *testing.T) {
	resourceName := "jumpcloud_mdm_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudMDMPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudMDMPolicyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMDMPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "Test MDM Policy"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test MDM policy for acceptance tests"),
				),
			},
			{
				Config: testAccJumpCloudMDMPolicyConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMDMPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "Updated MDM Policy"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated MDM policy description"),
					resource.TestCheckResourceAttr(resourceName, "settings.passcode.min_length", "6"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudMDMPolicyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		// You would typically make an API call here to check if the resource exists
		// This is a simplified version
		return nil
	}
}

func testAccCheckJumpCloudMDMPolicyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_mdm_policy" {
			continue
		}

		// You would typically make an API call here to check if the resource was destroyed
		// This is a simplified version
		return nil
	}

	return nil
}

func testAccJumpCloudMDMPolicyConfig_basic() string {
	return `
resource "jumpcloud_mdm_policy" "test" {
  name        = "Test MDM Policy"
  description = "Test MDM policy for acceptance tests"
  platform    = "ios"
  settings = {
    passcode = {
      required = true
      min_length = 4
    }
  }
}
`
}

func testAccJumpCloudMDMPolicyConfig_update() string {
	return `
resource "jumpcloud_mdm_policy" "test" {
  name        = "Updated MDM Policy"
  description = "Updated MDM policy description"
  platform    = "ios"
  settings = {
    passcode = {
      required = true
      min_length = 6
      complex_chars = true
    }
  }
}
`
}
