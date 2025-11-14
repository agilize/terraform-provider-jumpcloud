package devices

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudDataSourceSystemDevices_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceSystemDevicesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_system_devices.all", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_system_devices.all", "devices.#"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceSystemDevicesConfig_basic() string {
	return `
data "jumpcloud_system_devices" "all" {
}
`
}

func TestAccJumpCloudDataSourceSystemDevices_filtered(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceSystemDevicesConfig_filtered(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_system_devices.filtered", "id"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceSystemDevicesConfig_filtered() string {
	return `
data "jumpcloud_system_devices" "filtered" {
  filter {
    field    = "os"
    operator = "eq"
    value    = "linux"
  }
  
  sort {
    field     = "hostname"
    direction = "asc"
  }
}

output "linux_device_count" {
  value = length(data.jumpcloud_system_devices.filtered.devices)
}
`
}

func TestAccJumpCloudDataSourceSystemDevices_withTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceSystemDevicesConfig_withTags(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_system_devices.tagged", "id"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceSystemDevicesConfig_withTags() string {
	return `
data "jumpcloud_system_devices" "tagged" {
  filter {
    field    = "tags"
    operator = "contains"
    value    = "production"
  }
}

output "tagged_device_count" {
  value = length(data.jumpcloud_system_devices.tagged.devices)
}
`
}
