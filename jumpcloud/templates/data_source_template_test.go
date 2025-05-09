package templates

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// These would be defined in a shared testing.go file
// nolint:unused
var testAccProviders map[string]*schema.Provider

// nolint:unused
var testAccProvider *schema.Provider

func testAccPreCheck(t *testing.T) {
	// Implementation would be in testing.go
	commonTesting.AccPreCheck(t)
}

func TestAccDataSourceExample_byID(t *testing.T) {
	// Skip template tests for now
	t.Skip("Skipping template test - this is a template for actual tests")

	resourceName := "jumpcloud_example.test"
	dataSourceName := "data.jumpcloud_example.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckExampleDestroy,
		Steps: []resource.TestStep{
			// First create a resource
			{
				Config: testAccDataSourceExampleConfig_resourceOnly("example-data-resource"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExampleExists(resourceName),
				),
			},
			// Then test reading it by ID
			{
				Config: testAccDataSourceExampleConfig_byID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "type", resourceName, "type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
				),
			},
		},
	})
}

func TestAccDataSourceExample_byName(t *testing.T) {
	// Skip template tests for now
	t.Skip("Skipping template test - this is a template for actual tests")

	resourceName := "jumpcloud_example.test"
	dataSourceName := "data.jumpcloud_example.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckExampleDestroy,
		Steps: []resource.TestStep{
			// First create a resource
			{
				Config: testAccDataSourceExampleConfig_resourceOnly("example-data-resource"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExampleExists(resourceName),
				),
			},
			// Then test reading it by name
			{
				Config: testAccDataSourceExampleConfig_byName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "type", resourceName, "type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
				),
			},
		},
	})
}

func TestAccDataSourceExamples_noFilter(t *testing.T) {
	// Skip template tests for now
	t.Skip("Skipping template test - this is a template for actual tests")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckExampleDestroy,
		Steps: []resource.TestStep{
			// First create multiple resources
			{
				Config: testAccDataSourceExamplesConfig_multipleResources(),
			},
			// Then test reading all of them
			{
				Config: testAccDataSourceExamplesConfig_noFilter(),
				Check: resource.ComposeTestCheckFunc(
					// We can't know exactly how many resources will be returned if using real APIs
					// in acceptance tests, but we can check that the output contains the ones we created
					resource.TestCheckResourceAttrSet("data.jumpcloud_examples.all", "examples.#"),
					resource.TestCheckTypeSetElemNestedAttrs("data.jumpcloud_examples.all", "examples.*", map[string]string{
						"name": "example-list-1",
						"type": "type1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.jumpcloud_examples.all", "examples.*", map[string]string{
						"name": "example-list-2",
						"type": "type2",
					}),
				),
			},
		},
	})
}

func TestAccDataSourceExamples_withFilters(t *testing.T) {
	// Skip template tests for now
	t.Skip("Skipping template test - this is a template for actual tests")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckExampleDestroy,
		Steps: []resource.TestStep{
			// First create multiple resources
			{
				Config: testAccDataSourceExamplesConfig_multipleResources(),
			},
			// Then test filtering by type
			{
				Config: testAccDataSourceExamplesConfig_filterByType("type1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_examples.filtered", "examples.#"),
					resource.TestCheckTypeSetElemNestedAttrs("data.jumpcloud_examples.filtered", "examples.*", map[string]string{
						"name": "example-list-1",
						"type": "type1",
					}),
				),
			},
			// Then test filtering by name (partial match)
			{
				Config: testAccDataSourceExamplesConfig_filterByName("list-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.jumpcloud_examples.filtered", "examples.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("data.jumpcloud_examples.filtered", "examples.*", map[string]string{
						"name": "example-list-2",
						"type": "type2",
					}),
				),
			},
		},
	})
}

// Config generation functions for data source tests
func testAccDataSourceExampleConfig_resourceOnly(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_example" "test" {
  name        = "%s"
  description = "Used for data source testing"
  type        = "type1"
  status      = "active"
  tags        = ["test", "data-source"]
}
`, name)
}

func testAccDataSourceExampleConfig_byID() string {
	return `
resource "jumpcloud_example" "test" {
  name        = "example-data-resource"
  description = "Used for data source testing"
  type        = "type1"
  status      = "active"
  tags        = ["test", "data-source"]
}

data "jumpcloud_example" "test" {
  id = jumpcloud_example.test.id
}
`
}

func testAccDataSourceExampleConfig_byName() string {
	return `
resource "jumpcloud_example" "test" {
  name        = "example-data-resource"
  description = "Used for data source testing"
  type        = "type1"
  status      = "active"
  tags        = ["test", "data-source"]
}

data "jumpcloud_example" "test" {
  name = jumpcloud_example.test.name
}
`
}

func testAccDataSourceExamplesConfig_multipleResources() string {
	return `
resource "jumpcloud_example" "resource1" {
  name        = "example-list-1"
  description = "First example resource for list testing"
  type        = "type1"
  status      = "active"
}

resource "jumpcloud_example" "resource2" {
  name        = "example-list-2"
  description = "Second example resource for list testing"
  type        = "type2"
  status      = "active"
}
`
}

func testAccDataSourceExamplesConfig_noFilter() string {
	return testAccDataSourceExamplesConfig_multipleResources() + `
data "jumpcloud_examples" "all" {
  # No filter means return all examples
}
`
}

func testAccDataSourceExamplesConfig_filterByType(resourceType string) string {
	return testAccDataSourceExamplesConfig_multipleResources() + fmt.Sprintf(`
data "jumpcloud_examples" "filtered" {
  filter {
    type = "%s"
  }
}
`, resourceType)
}

func testAccDataSourceExamplesConfig_filterByName(namePattern string) string {
	return testAccDataSourceExamplesConfig_multipleResources() + fmt.Sprintf(`
data "jumpcloud_examples" "filtered" {
  filter {
    name = "%s"
  }
}
`, namePattern)
}
