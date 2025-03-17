package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceAdminUser_basic(t *testing.T) {
	var adminID string

	rEmail := acctest.RandomWithPrefix("admin-test") + "@example.com"
	resourceName := "jumpcloud_admin_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAdminUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAdminUserConfig_basic(rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAdminUserExists(resourceName, &adminID),
					resource.TestCheckResourceAttr(resourceName, "email", rEmail),
					resource.TestCheckResourceAttr(resourceName, "firstname", "Test"),
					resource.TestCheckResourceAttr(resourceName, "lastname", "Admin"),
					resource.TestCheckResourceAttr(resourceName, "is_super_admin", "false"),
					resource.TestMatchResourceAttr(resourceName, "created", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
					resource.TestMatchResourceAttr(resourceName, "updated", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccResourceAdminUser_update(t *testing.T) {
	var adminID string

	rEmail := acctest.RandomWithPrefix("admin-test") + "@example.com"
	resourceName := "jumpcloud_admin_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAdminUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAdminUserConfig_basic(rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAdminUserExists(resourceName, &adminID),
					resource.TestCheckResourceAttr(resourceName, "email", rEmail),
					resource.TestCheckResourceAttr(resourceName, "firstname", "Test"),
					resource.TestCheckResourceAttr(resourceName, "lastname", "Admin"),
				),
			},
			{
				Config: testAccAdminUserConfig_updated(rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAdminUserExists(resourceName, &adminID),
					resource.TestCheckResourceAttr(resourceName, "email", rEmail),
					resource.TestCheckResourceAttr(resourceName, "firstname", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "lastname", "AdminUser"),
					resource.TestCheckResourceAttr(resourceName, "is_super_admin", "true"),
				),
			},
		},
	})
}

func testAccCheckAdminUserExists(resourceName string, adminID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		*adminID = rs.Primary.ID

		return nil
	}
}

func testAccCheckAdminUserDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_admin_user" {
			continue
		}

		// Retrieve the client from the test provider
		client := testAccProvider.Meta().(JumpCloudClient)

		// Check that the admin user no longer exists
		url := fmt.Sprintf("/api/v2/administrators/%s", rs.Primary.ID)
		_, err := client.DoRequest("GET", url, nil)

		// The request should return an error if the admin user is destroyed
		if err == nil {
			return fmt.Errorf("JumpCloud admin user %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAdminUserConfig_basic(email string) string {
	return fmt.Sprintf(`
resource "jumpcloud_admin_user" "test" {
  email       = %q
  firstname   = "Test"
  lastname    = "Admin"
  password    = "TestPassword123!"
  is_super_admin = false
}
`, email)
}

func testAccAdminUserConfig_updated(email string) string {
	return fmt.Sprintf(`
resource "jumpcloud_admin_user" "test" {
  email       = %q
  firstname   = "Updated"
  lastname    = "AdminUser"
  password    = "UpdatedPassword123!"
  is_super_admin = true
}
`, email)
}
