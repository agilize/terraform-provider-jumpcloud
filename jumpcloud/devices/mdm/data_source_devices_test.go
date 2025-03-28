package mdm_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudDataSourceMDMDevices_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceMDMDevicesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_mdm_devices.all", "id"),
					// The following check may fail if there are no MDM devices, but is useful
					// if at least one device exists in the test environment
					resource.TestCheckResourceAttrSet("data.jumpcloud_mdm_devices.all", "devices.#"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceMDMDevicesConfig_basic() string {
	return `
data "jumpcloud_mdm_devices" "all" {
}
`
}

func TestAccJumpCloudDataSourceMDMDevices_filtered(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceMDMDevicesConfig_filtered(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_mdm_devices.ios", "id"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceMDMDevicesConfig_filtered() string {
	return `
data "jumpcloud_mdm_devices" "ios" {
  filter {
    field    = "platform"
    operator = "eq"
    value    = "ios"
  }
  
  sort {
    field     = "name"
    direction = "asc"
  }
}
`
}

func TestAccJumpCloudDataSourceMDMDevices_corporate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceMDMDevicesConfig_corporate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_mdm_devices.corporate", "id"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceMDMDevicesConfig_corporate() string {
	return `
data "jumpcloud_mdm_devices" "corporate" {
  filter {
    field    = "ownership"
    operator = "eq"
    value    = "corporate"
  }
}

output "corporate_device_count" {
  value = length(data.jumpcloud_mdm_devices.corporate.devices)
}
`
}
