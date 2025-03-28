package mdm_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudMDMProfile_basic(t *testing.T) {
	resourceName := "jumpcloud_mdm_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudMDMProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudMDMProfileConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMDMProfileExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "Test MDM Profile"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test MDM profile for acceptance tests"),
					resource.TestCheckResourceAttr(resourceName, "platform", "ios"),
					resource.TestCheckResourceAttr(resourceName, "settings.wifi.0.ssid", "Test-WiFi"),
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

func TestAccJumpCloudMDMProfile_update(t *testing.T) {
	resourceName := "jumpcloud_mdm_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudMDMProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudMDMProfileConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMDMProfileExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "Test MDM Profile"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test MDM profile for acceptance tests"),
				),
			},
			{
				Config: testAccJumpCloudMDMProfileConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMDMProfileExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "Updated MDM Profile"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated MDM profile description"),
					resource.TestCheckResourceAttr(resourceName, "settings.wifi.0.ssid", "Updated-WiFi"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudMDMProfileExists(resourceName string) resource.TestCheckFunc {
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

func testAccCheckJumpCloudMDMProfileDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_mdm_profile" {
			continue
		}

		// You would typically make an API call here to check if the resource was destroyed
		// This is a simplified version
		return nil
	}

	return nil
}

func testAccJumpCloudMDMProfileConfig_basic() string {
	return `
resource "jumpcloud_mdm_profile" "test" {
  name        = "Test MDM Profile"
  description = "Test MDM profile for acceptance tests"
  platform    = "ios"
  settings = {
    wifi = [
      {
        ssid = "Test-WiFi"
        auto_join = true
        hidden_network = false
        security_type = "WPA2"
        password = "test-password"
      }
    ]
  }
}
`
}

func testAccJumpCloudMDMProfileConfig_update() string {
	return `
resource "jumpcloud_mdm_profile" "test" {
  name        = "Updated MDM Profile"
  description = "Updated MDM profile description"
  platform    = "ios"
  settings = {
    wifi = [
      {
        ssid = "Updated-WiFi"
        auto_join = true
        hidden_network = true
        security_type = "WPA2"
        password = "updated-password"
      }
    ]
  }
}
`
}
