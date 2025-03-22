package appcatalog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/testing"
)

func TestAccDataSourceApplication_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceApplicationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_appcatalog_application.test", "name"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_appcatalog_application.test", "app_type"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_appcatalog_application.test", "status"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_appcatalog_application.test", "visibility"),
				),
			},
		},
	})
}

func testAccDataSourceApplicationConfig_basic() string {
	return `
# First create or find an application
resource "jumpcloud_appcatalog_application" "test_app" {
  name        = "Test Application"
  description = "Application for testing data source"
  app_type    = "web"
  status      = "active"
  visibility  = "private"
}

# Then retrieve it with the data source
data "jumpcloud_appcatalog_application" "test" {
  id = jumpcloud_appcatalog_application.test_app.id
}
`
}
