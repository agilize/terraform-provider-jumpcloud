package app_catalog

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// These variables would be declared in a common testing file
// and would be imported into this file
// var testAccProviders map[string]*schema.Provider
// var testAccProvider *schema.Provider
// func testAccPreCheck(t *testing.T) {}

func TestAccResourceCategory_basic(t *testing.T) {
	var resourceName = "jumpcloud_appcatalog_category.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCategoryConfig_basic("test-category"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-category"),
					resource.TestCheckResourceAttr(resourceName, "display_order", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceCategory_update(t *testing.T) {
	var resourceName = "jumpcloud_appcatalog_category.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCategoryConfig_basic("initial-name"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "initial-name"),
				),
			},
			{
				Config: testAccResourceCategoryConfig_update("updated-name", "Updated description", 10),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "updated-name"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "display_order", "10"),
				),
			},
		},
	})
}

func TestAccResourceCategory_applications(t *testing.T) {
	var resourceName = "jumpcloud_appcatalog_category.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCategoryConfig_applications("category-with-apps"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "category-with-apps"),
					resource.TestCheckResourceAttr(resourceName, "applications.#", "2"),
				),
			},
		},
	})
}

func testAccResourceCategoryConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_appcatalog_category" "test" {
  name = "%s"
}
`, name)
}

func testAccResourceCategoryConfig_update(name, description string, displayOrder int) string {
	return fmt.Sprintf(`
resource "jumpcloud_appcatalog_category" "test" {
  name         = "%s"
  description  = "%s"
  display_order = %d
}
`, name, description, displayOrder)
}

func testAccResourceCategoryConfig_applications(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_appcatalog_application" "app1" {
  name        = "App 1"
  description = "Application 1 for category test"
  app_type    = "web"
  status      = "active"
}

resource "jumpcloud_appcatalog_application" "app2" {
  name        = "App 2"
  description = "Application 2 for category test"
  app_type    = "web"
  status      = "active"
}

resource "jumpcloud_appcatalog_category" "test" {
  name        = "%s"
  applications = [
    jumpcloud_appcatalog_application.app1.id,
    jumpcloud_appcatalog_application.app2.id
  ]
}
`, name)
}
