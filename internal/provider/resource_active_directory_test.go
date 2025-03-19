package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceActiveDirectory_basic(t *testing.T) {
	var adID string

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_active_directory.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckActiveDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccActiveDirectoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActiveDirectoryExists(resourceName, &adID),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "domain", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "type", "regular"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "use_ou", "false"),
					resource.TestMatchResourceAttr(resourceName, "created", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
					resource.TestMatchResourceAttr(resourceName, "updated", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
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

func TestAccResourceActiveDirectory_update(t *testing.T) {
	var adID string

	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := rName + "-updated"
	resourceName := "jumpcloud_active_directory.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckActiveDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccActiveDirectoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActiveDirectoryExists(resourceName, &adID),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "domain", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "description", "Basic Active Directory integration"),
				),
			},
			{
				Config: testAccActiveDirectoryConfig_updated(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActiveDirectoryExists(resourceName, &adID),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated Active Directory integration"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "use_ou", "true"),
					resource.TestCheckResourceAttr(resourceName, "ou_path", "OU=JumpCloud,DC=example,DC=com"),
				),
			},
		},
	})
}

func testAccCheckActiveDirectoryExists(resourceName string, adID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		*adID = rs.Primary.ID

		return nil
	}
}

func testAccCheckActiveDirectoryDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_active_directory" {
			continue
		}

		// Retrieve the client from the test provider
		client := testAccProvider.Meta().(JumpCloudClient)

		// Check that the Active Directory integration no longer exists
		url := fmt.Sprintf("/api/v2/activedirectories/%s", rs.Primary.ID)
		_, err := client.DoRequest("GET", url, nil)

		// The request should return an error if the Active Directory is destroyed
		if err == nil {
			return fmt.Errorf("JumpCloud Active Directory integration %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccActiveDirectoryConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_active_directory" "test" {
  name        = %q
  description = "Basic Active Directory integration"
  domain      = "example.com"
  type        = "regular"
  enabled     = true
  use_ou      = false
}
`, rName)
}

func testAccActiveDirectoryConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_active_directory" "test" {
  name        = %q
  description = "Updated Active Directory integration"
  domain      = "example.com"
  type        = "regular"
  enabled     = false
  use_ou      = true
  ou_path     = "OU=JumpCloud,DC=example,DC=com"
}
`, rName)
}
