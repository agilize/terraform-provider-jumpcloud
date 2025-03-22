package appcatalog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/testing"
)

func TestAccResourceAppCatalogApplication_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAppCatalogApplicationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("jumpcloud_appcatalog_application.test", "id"),
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_application.test", "name", "Test Application"),
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_application.test", "description", "Application for testing"),
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_application.test", "app_type", "web"),
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_application.test", "status", "active"),
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_application.test", "visibility", "private"),
				),
			},
		},
	})
}

func TestAccResourceAppCatalogApplication_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAppCatalogApplicationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_application.test", "name", "Test Application"),
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_application.test", "description", "Application for testing"),
				),
			},
			{
				Config: testAccResourceAppCatalogApplicationConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_application.test", "name", "Updated Test Application"),
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_application.test", "description", "Updated application for testing"),
				),
			},
		},
	})
}

func testAccResourceAppCatalogApplicationConfig_basic() string {
	return `
resource "jumpcloud_appcatalog_application" "test" {
  name        = "Test Application"
  description = "Application for testing"
  app_type    = "web"
  status      = "active"
  visibility  = "private"
}
`
}

func testAccResourceAppCatalogApplicationConfig_updated() string {
	return `
resource "jumpcloud_appcatalog_application" "test" {
  name        = "Updated Test Application"
  description = "Updated application for testing"
  app_type    = "web"
  status      = "active"
  visibility  = "private"
}
`
}
