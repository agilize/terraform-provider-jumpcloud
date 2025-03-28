package app_catalog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccDataSourceAppCatalogCategories_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAppCatalogCategoriesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_appcatalog_categories.test", "categories.#"),
				),
			},
		},
	})
}

func testAccDataSourceAppCatalogCategoriesConfig_basic() string {
	return `
data "jumpcloud_appcatalog_categories" "test" {
}
`
}
