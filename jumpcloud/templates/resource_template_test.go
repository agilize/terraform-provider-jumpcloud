package templates

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccResourceExample_basic(t *testing.T) {
	// Skip template tests for now
	t.Skip("Skipping template test - this is a template for actual tests")

	var resourceName = "jumpcloud_example.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckExampleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceExampleConfig_basic("example-resource-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExampleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "example-resource-name"),
					resource.TestCheckResourceAttr(resourceName, "type", "type1"),
					resource.TestCheckResourceAttr(resourceName, "status", "active"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Don't verify sensitive fields that aren't returned by the API
				ImportStateVerifyIgnore: []string{"api_token"},
			},
		},
	})
}

func TestAccResourceExample_update(t *testing.T) {
	// Skip template tests for now
	t.Skip("Skipping template test - this is a template for actual tests")

	var resourceName = "jumpcloud_example.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckExampleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceExampleConfig_basic("example-resource-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExampleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "example-resource-name"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				Config: testAccResourceExampleConfig_update("example-resource-name", "Updated description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExampleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "example-resource-name"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccResourceExample_tags(t *testing.T) {
	// Skip template tests for now
	t.Skip("Skipping template test - this is a template for actual tests")

	var resourceName = "jumpcloud_example.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckExampleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceExampleConfig_tags("example-tags"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExampleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "example-tags"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
				),
			},
			{
				Config: testAccResourceExampleConfig_tagsUpdate("example-tags"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExampleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "example-tags"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "tags.2", "tag3"),
				),
			},
		},
	})
}

func TestAccResourceExample_validation(t *testing.T) {
	// Skip template tests for now
	t.Skip("Skipping template test - this is a template for actual tests")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckExampleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceExampleConfig_invalidType("example-invalid"),
				ExpectError: regexp.MustCompile(`expected type to be one of \[type1 type2 type3\]`),
			},
		},
	})
}

func testAccCheckExampleExists(n string) resource.TestCheckFunc {
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
		// resp, err := client.DoRequest("GET", fmt.Sprintf("/api/v2/example/resources/%s", rs.Primary.ID), nil)
		// if err != nil {
		//    return fmt.Errorf("error fetching item with ID %s: %s", rs.Primary.ID, err)
		// }

		return nil
	}
}

func testAccCheckExampleDestroy(s *terraform.State) error {
	// Add API checking logic here if you want to validate that the resource was destroyed
	// For example:
	// client := testAccProvider.Meta().(*apiclient.Client)
	//
	// for _, rs := range s.RootModule().Resources {
	//    if rs.Type != "jumpcloud_example" {
	//       continue
	//    }
	//
	//    _, err := client.DoRequest("GET", fmt.Sprintf("/api/v2/example/resources/%s", rs.Primary.ID), nil)
	//    if err == nil {
	//       return fmt.Errorf("Example resource %s still exists", rs.Primary.ID)
	//    }
	// }

	return nil
}

// Config generation functions with clear naming conventions
func testAccResourceExampleConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_example" "test" {
  name   = "%s"
  type   = "type1"
  status = "active"
}
`, name)
}

func testAccResourceExampleConfig_update(name, description string) string {
	return fmt.Sprintf(`
resource "jumpcloud_example" "test" {
  name        = "%s"
  description = "%s"
  type        = "type1"
  status      = "active"
}
`, name, description)
}

func testAccResourceExampleConfig_tags(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_example" "test" {
  name   = "%s"
  type   = "type1"
  status = "active"
  tags   = ["tag1", "tag2"]
}
`, name)
}

func testAccResourceExampleConfig_tagsUpdate(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_example" "test" {
  name   = "%s"
  type   = "type1"
  status = "active"
  tags   = ["tag1", "tag2", "tag3"]
}
`, name)
}

func testAccResourceExampleConfig_invalidType(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_example" "test" {
  name   = "%s"
  type   = "invalid-type"
  status = "active"
}
`, name)
}
