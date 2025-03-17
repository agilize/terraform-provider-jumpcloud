package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSystemMetrics_basic(t *testing.T) {
	dataSourceName := "data.jumpcloud_system_metrics.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSystemMetricsConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "metrics.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "total"),
				),
			},
		},
	})
}

func TestAccDataSourceSystemMetrics_filtered(t *testing.T) {
	dataSourceName := "data.jumpcloud_system_metrics.filtered"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSystemMetricsConfig_filtered(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "metrics.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "total"),
				),
			},
		},
	})
}

func TestAccDataSourceSystemMetrics_pagination(t *testing.T) {
	dataSourceName := "data.jumpcloud_system_metrics.paginated"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSystemMetricsConfig_pagination(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "metrics.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "total"),
				),
			},
		},
	})
}

func testAccDataSourceSystemMetricsConfig_basic() string {
	return `
data "jumpcloud_system_metrics" "test" {}

output "all_metrics" {
  value = data.jumpcloud_system_metrics.test.metrics
}

output "metrics_total" {
  value = data.jumpcloud_system_metrics.test.total
}
`
}

func testAccDataSourceSystemMetricsConfig_filtered() string {
	return `
data "jumpcloud_system_metrics" "filtered" {
  os = "linux"
  start_date = "2023-01-01T00:00:00Z"
  end_date = "2023-12-31T23:59:59Z"
}

output "filtered_metrics" {
  value = data.jumpcloud_system_metrics.filtered.metrics
}

output "filtered_total" {
  value = data.jumpcloud_system_metrics.filtered.total
}
`
}

func testAccDataSourceSystemMetricsConfig_pagination() string {
	return `
data "jumpcloud_system_metrics" "paginated" {
  limit = 10
  skip = 0
  sort = "timestamp"
  sort_dir = "DESC"
}

output "paginated_metrics" {
  value = data.jumpcloud_system_metrics.paginated.metrics
}

output "paginated_total" {
  value = data.jumpcloud_system_metrics.paginated.total
}
`
}
