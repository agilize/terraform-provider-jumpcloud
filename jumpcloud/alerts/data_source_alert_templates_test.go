package alerts

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/testing"
)

// TestDataSourceAlertTemplatesSchema tests the schema structure of the alert templates data source
func TestDataSourceAlertTemplatesSchema(t *testing.T) {
	s := DataSourceAlertTemplates()

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

	// Test templates field
	if s.Schema["templates"] == nil {
		t.Error("Expected templates in schema, but it does not exist")
	}
	if s.Schema["templates"].Type != schema.TypeList {
		t.Error("Expected templates to be of type list")
	}
	if !s.Schema["templates"].Computed {
		t.Error("Expected templates to be computed")
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
func TestAccDataSourceAlertTemplates_basic(t *testing.T) {
	dataSourceName := "data.jumpcloud_alert_templates.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceAlertTemplatesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "total_count"),
				),
			},
		},
	})
}

func TestAccDataSourceAlertTemplates_filtered(t *testing.T) {
	dataSourceName := "data.jumpcloud_alert_templates.filtered"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceAlertTemplatesConfig_filtered(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "total_count"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceAlertTemplatesConfig_basic() string {
	return `
data "jumpcloud_alert_templates" "test" {
}
`
}

func testAccJumpCloudDataSourceAlertTemplatesConfig_filtered() string {
	return `
data "jumpcloud_alert_templates" "filtered" {
  filter {
    type = "system_metric"
    category = "operations"
    pre_configured = true
  }
  
  sort {
    field = "name"
    direction = "asc"
  }
  
  limit = 10
}
`
}
