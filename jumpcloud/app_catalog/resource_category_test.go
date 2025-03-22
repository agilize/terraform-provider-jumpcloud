package appcatalog

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/testing"
)

// These variables would be declared in a common testing file
// and would be imported into this file
// var testAccProviders map[string]*schema.Provider
// var testAccProvider *schema.Provider
// func testAccPreCheck(t *testing.T) {}

func TestAccResourceCategory_basic(t *testing.T) {
	var resourceName = "jumpcloud_appcatalog_category.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
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
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCategoryConfig_basic("test-category"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-category"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				Config: testAccResourceCategoryConfig_update("test-category", "Updated description", 5),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-category"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "display_order", "5"),
				),
			},
		},
	})
}

func TestAccResourceCategory_applications(t *testing.T) {
	var resourceName = "jumpcloud_appcatalog_category.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCategoryConfig_applications("test-category-apps"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-category-apps"),
					resource.TestCheckResourceAttr(resourceName, "applications.#", "0"),
				),
			},
			// In a real test we would add applications here, but that requires
			// existing application IDs which would be difficult in an isolated test
		},
	})
}

// Helper functions
// These would be implemented once the testing utilities are set up

/*
func testAccCheckCategoryExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		// Add API checking logic here if you want to validate from the API
		// For example:
		// client := testAccProvider.Meta().(*apiclient.Client)
		// resp, err := client.DoRequest("GET", fmt.Sprintf("/api/v2/appcatalog/categories/%s", rs.Primary.ID), nil)
		// if err != nil {
		//    return fmt.Errorf("error fetching category with ID %s: %s", rs.Primary.ID, err)
		// }

		return nil
	}
}

func testAccCheckCategoryDestroy(s *terraform.State) error {
	// Add API checking logic here if you want to validate that the resource was destroyed
	// For example:
	// client := testAccProvider.Meta().(*apiclient.Client)
	//
	// for _, rs := range s.RootModule().Resources {
	//    if rs.Type != "jumpcloud_appcatalog_category" {
	//       continue
	//    }
	//
	//    _, err := client.DoRequest("GET", fmt.Sprintf("/api/v2/appcatalog/categories/%s", rs.Primary.ID), nil)
	//    if err == nil {
	//       return fmt.Errorf("Category %s still exists", rs.Primary.ID)
	//    }
	// }

	return nil
}
*/

// Config generation functions
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
  name          = "%s"
  description   = "%s"
  display_order = %d
}
`, name, description, displayOrder)
}

func testAccResourceCategoryConfig_applications(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_appcatalog_category" "test" {
  name = "%s"
  # Applications would be added here in a real test
  # applications = ["app-id-1", "app-id-2"]
}
`, name)
}
