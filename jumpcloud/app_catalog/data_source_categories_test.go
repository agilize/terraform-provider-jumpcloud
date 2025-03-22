package appcatalog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAppCatalogCategories_basic(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	dataSourceName := "data.jumpcloud_app_catalog_categories.all"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAppCatalogCategoriesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "categories.#"),
				),
			},
		},
	})
}

func testAccDataSourceAppCatalogCategoriesConfig_basic() string {
	return `
data "jumpcloud_app_catalog_categories" "all" {}
`
}
