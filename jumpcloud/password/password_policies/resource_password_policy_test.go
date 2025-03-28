package password_policies

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// Definindo as provider factories

func TestAccResourcePasswordPolicy_basic(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_password_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { commonTesting.TestAccPreCheck(t) },
		Providers:    commonTesting.TestAccProviders,
		CheckDestroy: testAccCheckPasswordPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPasswordPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPasswordPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "active"),
					resource.TestCheckResourceAttr(resourceName, "min_length", "10"),
					resource.TestCheckResourceAttr(resourceName, "require_uppercase", "true"),
					resource.TestCheckResourceAttr(resourceName, "require_lowercase", "true"),
					resource.TestCheckResourceAttr(resourceName, "require_number", "true"),
					resource.TestCheckResourceAttr(resourceName, "require_symbol", "true"),
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

// Definindo as provider factories

func TestAccResourcePasswordPolicy_update(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := rName + "-updated"
	resourceName := "jumpcloud_password_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { commonTesting.TestAccPreCheck(t) },
		Providers:    commonTesting.TestAccProviders,
		CheckDestroy: testAccCheckPasswordPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPasswordPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPasswordPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Basic password policy"),
					resource.TestCheckResourceAttr(resourceName, "min_length", "10"),
				),
			},
			{
				Config: testAccPasswordPolicyConfig_updated(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPasswordPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated password policy"),
					resource.TestCheckResourceAttr(resourceName, "min_length", "12"),
					resource.TestCheckResourceAttr(resourceName, "minimum_age", "7"),
					resource.TestCheckResourceAttr(resourceName, "disallow_previous_passwords", "5"),
				),
			},
		},
	})
}

// Definindo as provider factories

func testAccCheckPasswordPolicyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		return nil
	}
}

// Definindo as provider factories

func testAccCheckPasswordPolicyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_password_policy" {
			continue
		}

		// Check that the password policy no longer exists
		// This would typically involve making an API call to verify

		return nil
	}

	return nil
}

// Definindo as provider factories

func testAccPasswordPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_password_policy" "test" {
  name        = %q
  description = "Basic password policy"
  status      = "active"
  min_length  = 10
  max_length  = 64
  require_uppercase = true
  require_lowercase = true
  require_number = true
  require_symbol = true
  expiration_time = 90
  expiration_warning_time = 14
  scope = "organization"
}
`, rName)
}

// Definindo as provider factories

func testAccPasswordPolicyConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_password_policy" "test" {
  name        = %q
  description = "Updated password policy"
  status      = "active"
  min_length  = 12
  max_length  = 64
  require_uppercase = true
  require_lowercase = true
  require_number = true
  require_symbol = true
  minimum_age = 7
  expiration_time = 60
  expiration_warning_time = 7
  disallow_previous_passwords = 5
  disallow_common_passwords = true
  scope = "organization"
}
`, rName)
}
