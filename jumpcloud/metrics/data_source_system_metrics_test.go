package metrics

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/testing"
)

// TestDataSourceSystemMetricsSchema tests the schema structure of the system metrics data source
func TestDataSourceSystemMetricsSchema(t *testing.T) {
	s := DataSourceSystemMetrics()

	// Test filter field
	if s.Schema["filter"] == nil {
		t.Error("Expected filter in schema, but it does not exist")
	}
	if s.Schema["filter"].Type != schema.TypeList {
		t.Error("Expected filter to be of type list")
	}
	if s.Schema["filter"].Required {
		t.Error("Expected filter to be optional")
	}

	// Test aggregation field
	if s.Schema["aggregation"] == nil {
		t.Error("Expected aggregation in schema, but it does not exist")
	}
	if s.Schema["aggregation"].Type != schema.TypeList {
		t.Error("Expected aggregation to be of type list")
	}
	if s.Schema["aggregation"].Required {
		t.Error("Expected aggregation to be optional")
	}

	// Test sort field
	if s.Schema["sort"] == nil {
		t.Error("Expected sort in schema, but it does not exist")
	}
	if s.Schema["sort"].Type != schema.TypeList {
		t.Error("Expected sort to be of type list")
	}
	if s.Schema["sort"].Required {
		t.Error("Expected sort to be optional")
	}

	// Test limit field
	if s.Schema["limit"] == nil {
		t.Error("Expected limit in schema, but it does not exist")
	}
	if s.Schema["limit"].Type != schema.TypeInt {
		t.Error("Expected limit to be of type int")
	}
	if s.Schema["limit"].Required {
		t.Error("Expected limit to be optional")
	}

	// Test metrics field
	if s.Schema["metrics"] == nil {
		t.Error("Expected metrics in schema, but it does not exist")
	}
	if s.Schema["metrics"].Type != schema.TypeList {
		t.Error("Expected metrics to be of type list")
	}
	if !s.Schema["metrics"].Computed {
		t.Error("Expected metrics to be computed")
	}

	// Test total_count field
	if s.Schema["total_count"] == nil {
		t.Error("Expected total_count in schema, but it does not exist")
	}
	if s.Schema["total_count"].Type != schema.TypeInt {
		t.Error("Expected total_count to be of type int")
	}
	if !s.Schema["total_count"].Computed {
		t.Error("Expected total_count to be computed")
	}
}

// Acceptance testing
func TestAccDataSourceSystemMetrics_basic(t *testing.T) {
	dataSourceName := "data.jumpcloud_system_metrics.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceSystemMetricsConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "total_count"),
				),
			},
		},
	})
}

func TestAccDataSourceSystemMetrics_filtered(t *testing.T) {
	dataSourceName := "data.jumpcloud_system_metrics.filtered"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceSystemMetricsConfig_filtered(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "total_count"),
				),
			},
		},
	})
}

func TestAccDataSourceSystemMetrics_aggregated(t *testing.T) {
	dataSourceName := "data.jumpcloud_system_metrics.aggregated"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceSystemMetricsConfig_aggregated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "total_count"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceSystemMetricsConfig_basic() string {
	return `
data "jumpcloud_system_metrics" "test" {
}
`
}

func testAccJumpCloudDataSourceSystemMetricsConfig_filtered() string {
	return `
data "jumpcloud_system_metrics" "filtered" {
  filter {
    metric_type = "cpu"
    start_time = "2023-01-01T00:00:00Z"
    end_time = "2023-01-02T00:00:00Z"
  }
  
  sort {
    field = "timestamp"
    direction = "desc"
  }
  
  limit = 10
}
`
}

func testAccJumpCloudDataSourceSystemMetricsConfig_aggregated() string {
	return `
data "jumpcloud_system_metrics" "aggregated" {
  filter {
    metric_type = "cpu"
  }
  
  aggregation {
    function = "avg"
    interval = "5m"
  }
  
  limit = 20
}
`
}
