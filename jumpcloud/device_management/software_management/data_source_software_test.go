package software_management_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudDataSourceSoftware_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceSoftwareConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_software.all", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_software.all", "software.#"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceSoftwareConfig_basic() string {
	return `
data "jumpcloud_software" "all" {
}
`
}

func TestAccJumpCloudDataSourceSoftware_filtered(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceSoftwareConfig_filtered(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_software.filtered", "id"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceSoftwareConfig_filtered() string {
	return `
data "jumpcloud_software" "filtered" {
  filter {
    field    = "type"
    operator = "eq"
    value    = "application"
  }
  
  sort {
    field     = "name"
    direction = "asc"
  }
}

output "application_count" {
  value = length(data.jumpcloud_software.filtered.software)
}
`
}

func TestAccJumpCloudDataSourceSoftware_byStatus(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceSoftwareConfig_byStatus(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_software.active", "id"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceSoftwareConfig_byStatus() string {
	return `
data "jumpcloud_software" "active" {
  filter {
    field    = "status"
    operator = "eq"
    value    = "active"
  }
}

output "active_software_count" {
  value = length(data.jumpcloud_software.active.software)
}
`
}
