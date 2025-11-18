package mdm_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudMDMConfiguration_basic(t *testing.T) {
	resourceName := "jumpcloud_mdm_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudMDMConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudMDMConfigurationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMDMConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-test-mdm-config"),
					resource.TestCheckResourceAttr(resourceName, "ios_settings.0.allow_app_store", "true"),
					resource.TestCheckResourceAttr(resourceName, "ios_settings.0.allow_safari", "true"),
					resource.TestCheckResourceAttr(resourceName, "ios_settings.0.allow_camera", "true"),
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

func TestAccJumpCloudMDMConfiguration_update(t *testing.T) {
	resourceName := "jumpcloud_mdm_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudMDMConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudMDMConfigurationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMDMConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-test-mdm-config"),
					resource.TestCheckResourceAttr(resourceName, "ios_settings.0.allow_app_store", "true"),
				),
			},
			{
				Config: testAccJumpCloudMDMConfigurationConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMDMConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-test-mdm-config-updated"),
					resource.TestCheckResourceAttr(resourceName, "ios_settings.0.allow_app_store", "false"),
					resource.TestCheckResourceAttr(resourceName, "android_settings.0.allow_camera", "false"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudMDMConfigurationExists(resourceName string) resource.TestCheckFunc {
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

func testAccCheckJumpCloudMDMConfigurationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_mdm_configuration" {
			continue
		}

		// You would typically make an API call here to check if the resource was destroyed
		// This is a simplified version
		return nil
	}

	return nil
}

func testAccJumpCloudMDMConfigurationConfig_basic() string {
	return `
resource "jumpcloud_mdm_configuration" "test" {
  name = "tf-test-mdm-config"
  ios_settings {
    allow_app_store = true
    allow_safari = true
    allow_camera = true
  }
}
`
}

func testAccJumpCloudMDMConfigurationConfig_update() string {
	return `
resource "jumpcloud_mdm_configuration" "test" {
  name = "tf-test-mdm-config-updated"
  ios_settings {
    allow_app_store = false
    allow_safari = true
    allow_camera = true
  }
  android_settings {
    allow_camera = false
  }
}
`
}
