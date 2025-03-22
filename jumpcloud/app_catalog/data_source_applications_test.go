package appcatalog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/testing"
)

func TestAccDataSourceAppCatalogApplications_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAppCatalogApplicationsConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_appcatalog_applications.test", "applications.#"),
				),
			},
		},
	})
}

func TestAccDataSourceAppCatalogApplications_filtered(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAppCatalogApplicationsConfig_filtered(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_appcatalog_applications.test_filtered", "applications.#"),
				),
			},
		},
	})
}

func testAccDataSourceAppCatalogApplicationsConfig_basic() string {
	return `
data "jumpcloud_appcatalog_applications" "test" {
}
`
}

func testAccDataSourceAppCatalogApplicationsConfig_filtered() string {
	return `
data "jumpcloud_appcatalog_applications" "test_filtered" {
  filter {
    app_type = "web"
    status = "active"
  }
}
`
}
