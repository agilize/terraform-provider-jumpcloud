package radius_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudDataSourceRadius_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceRadiusConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_radius.all", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_radius.all", "radius_servers.#"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceRadiusConfig_basic() string {
	return `
data "jumpcloud_radius" "all" {
}
`
}

func TestAccJumpCloudDataSourceRadius_filtered(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceRadiusConfig_filtered(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_radius.filtered", "id"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceRadiusConfig_filtered() string {
	return `
data "jumpcloud_radius" "filtered" {
  filter {
    field    = "status"
    operator = "eq"
    value    = "active"
  }
  
  sort {
    field     = "name"
    direction = "asc"
  }
}

output "active_radius_count" {
  value = length(data.jumpcloud_radius.filtered.radius_servers)
}
`
}

func TestAccJumpCloudDataSourceRadius_withServer(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceRadiusConfig_withServer(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_radius.with_server", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_radius.with_server", "radius_servers.#"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceRadiusConfig_withServer() string {
	return `
resource "jumpcloud_radius_server" "test" {
  name        = "Test RADIUS Server"
  description = "Test RADIUS server created for data source test"
  host        = "radius.example.com"
  port        = 1812
  secret      = "test_secret"
  status      = "active"
}

data "jumpcloud_radius" "with_server" {
  filter {
    field    = "name"
    operator = "eq"
    value    = jumpcloud_radius_server.test.name
  }
}

output "found_server_id" {
  value = length(data.jumpcloud_radius.with_server.radius_servers) > 0 ? data.jumpcloud_radius.with_server.radius_servers[0].id : "not_found"
}
`
}
