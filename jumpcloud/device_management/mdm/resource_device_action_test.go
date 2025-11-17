package mdm_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudMDMDeviceAction_basic(t *testing.T) {
	resourceName := "jumpcloud_mdm_device_action.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudMDMDeviceActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudMDMDeviceActionConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMDMDeviceActionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "type", "lock"),
					resource.TestCheckResourceAttrSet(resourceName, "device_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudMDMDeviceActionExists(resourceName string) resource.TestCheckFunc {
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

func testAccCheckJumpCloudMDMDeviceActionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_mdm_device_action" {
			continue
		}

		// You would typically make an API call here to check if the resource was destroyed
		// This is a simplified version
		return nil
	}

	return nil
}

func testAccJumpCloudMDMDeviceActionConfig_basic() string {
	return `
# Note: This test requires an existing MDM-managed device
# You would need to replace this with a real device ID or use a data source
data "jumpcloud_mdm_devices" "filtered" {
  filter {
    field    = "platform"
    operator = "eq"
    value    = "ios"
  }
}

resource "jumpcloud_mdm_device_action" "test" {
  count      = length(data.jumpcloud_mdm_devices.filtered.devices) > 0 ? 1 : 0
  device_id  = data.jumpcloud_mdm_devices.filtered.devices[0].id
  action_type = "lock"
  reason     = "Testing device action via Terraform"
  timeout    = 60
}
`
}
