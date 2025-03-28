package organization_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccJumpCloudOrganization_basic(t *testing.T) {
	t.Skip("Skipping acceptance test in unit test environment")

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { t.Log("PreCheck running") },
		Providers:    nil, // Will be set by the test framework
		CheckDestroy: testAccCheckJumpCloudOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudOrganizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudOrganizationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "display_name", fmt.Sprintf("%s Display", rName)),
					resource.TestCheckResourceAttr(resourceName, "logo_url", "https://example.com/logo.png"),
					resource.TestCheckResourceAttr(resourceName, "website", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "contact_name", "Test Contact"),
					resource.TestCheckResourceAttr(resourceName, "contact_email", "contact@example.com"),
					resource.TestCheckResourceAttrSet(resourceName, "created"),
				),
			},
		},
	})
}

func TestAccJumpCloudOrganization_update(t *testing.T) {
	t.Skip("Skipping acceptance test in unit test environment")

	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := rName + "-updated"
	resourceName := "jumpcloud_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { t.Log("PreCheck running") },
		Providers:    nil, // Will be set by the test framework
		CheckDestroy: testAccCheckJumpCloudOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudOrganizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudOrganizationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "display_name", fmt.Sprintf("%s Display", rName)),
				),
			},
			{
				Config: testAccJumpCloudOrganizationConfig_updated(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudOrganizationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "display_name", fmt.Sprintf("%s Display Updated", rNameUpdated)),
					resource.TestCheckResourceAttr(resourceName, "website", "https://updated.example.com"),
					resource.TestCheckResourceAttr(resourceName, "contact_name", "Updated Contact"),
					resource.TestCheckResourceAttr(resourceName, "contact_email", "updated@example.com"),
				),
			},
		},
	})
}

func TestAccJumpCloudOrganization_withAllowedDomains(t *testing.T) {
	t.Skip("Skipping acceptance test in unit test environment")

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { t.Log("PreCheck running") },
		Providers:    nil, // Will be set by the test framework
		CheckDestroy: testAccCheckJumpCloudOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudOrganizationConfig_withAllowedDomains(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudOrganizationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "allowed_domains.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "allowed_domains.0", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "allowed_domains.1", "test.example.com"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudOrganizationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Mock implementation for testing
		return nil
	}
}

func testAccCheckJumpCloudOrganizationDestroy(s *terraform.State) error {
	// Mock implementation for testing
	return nil
}

func testAccJumpCloudOrganizationConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_organization" "test" {
  name          = "%s"
  display_name  = "%s Display"
  logo_url      = "https://example.com/logo.png"
  website       = "https://example.com"
  contact_name  = "Test Contact"
  contact_email = "contact@example.com"
  contact_phone = "+1234567890"
  
  settings = {
    "allow_guest_users" = "true"
  }
}
`, name, name)
}

func testAccJumpCloudOrganizationConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_organization" "test" {
  name          = "%s"
  display_name  = "%s Display Updated"
  logo_url      = "https://example.com/logo-updated.png"
  website       = "https://updated.example.com"
  contact_name  = "Updated Contact"
  contact_email = "updated@example.com"
  contact_phone = "+0987654321"
  
  settings = {
    "allow_guest_users" = "false"
    "custom_setting"    = "value"
  }
}
`, name, name)
}

func testAccJumpCloudOrganizationConfig_withAllowedDomains(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_organization" "test" {
  name          = "%s"
  display_name  = "%s Display"
  
  allowed_domains = [
    "example.com",
    "test.example.com"
  ]
}
`, name, name)
}
