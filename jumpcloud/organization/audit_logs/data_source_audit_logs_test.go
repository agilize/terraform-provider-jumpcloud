package audit_logs

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccDataSourceAdminAuditLogs_basic(t *testing.T) {
	dataSourceName := "data.jumpcloud_admin_audit_logs.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAdminAuditLogsConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "logs.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "total_count"),
					// We can add more specific checks if needed, but since this is
					// a data source that returns logs which may change frequently,
					// we're just verifying that the data source itself works.
				),
			},
		},
	})
}

func TestAccDataSourceAdminAuditLogs_filtered(t *testing.T) {
	dataSourceName := "data.jumpcloud_admin_audit_logs.filtered"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAdminAuditLogsConfig_filtered(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "logs.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "total_count"),
					// Add specific checks for the filtered data source
				),
			},
		},
	})
}

func testAccDataSourceAdminAuditLogsConfig_basic() string {
	return `
data "jumpcloud_admin_audit_logs" "test" {}
`
}

func testAccDataSourceAdminAuditLogsConfig_filtered() string {
	return `
data "jumpcloud_admin_audit_logs" "filtered" {
  filter {
    action = "login"
    success = true
  }
}
`
}
