package policies

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// Definindo as provider factories

func TestResourcePolicy(t *testing.T) {
	r := ResourcePolicy()
	// Use standard Go testing instead of assert
	if r == nil {
		t.Fatal("Expected non-nil resource")
	}
	if r.Schema["name"] == nil {
		t.Fatal("Expected non-nil name schema")
	}
	if r.Schema["type"] == nil {
		t.Fatal("Expected non-nil type schema")
	}
	if r.Schema["settings"] == nil {
		t.Fatal("Expected non-nil settings schema")
	}
	if r.CreateContext == nil {
		t.Fatal("Expected non-nil CreateContext")
	}
	if r.ReadContext == nil {
		t.Fatal("Expected non-nil ReadContext")
	}
	if r.UpdateContext == nil {
		t.Fatal("Expected non-nil UpdateContext")
	}
	if r.DeleteContext == nil {
		t.Fatal("Expected non-nil DeleteContext")
	}
}

// TestAccResourcePolicy tests the authentication policy resource
// Definindo as provider factories

func TestAccResourcePolicy(t *testing.T) {
	t.Skip("Skipping until CI environment is set up")

	resourceName := "jumpcloud_auth_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { jctest.TestAccPreCheck(t) },
		Providers:    jctest.TestAccProviders,
		CheckDestroy: testAccCheckJumpCloudAuthPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudAuthPolicyConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudAuthPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-test-auth-policy"),
					resource.TestCheckResourceAttr(resourceName, "disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "type", "user_portal"),
				),
			},
		},
	})
}

// Definindo as provider factories

func testAccCheckJumpCloudAuthPolicyDestroy(s *terraform.State) error {
	// Implementation depends on the test setup
	// This is a placeholder for the actual implementation
	return nil
}

// Definindo as provider factories

func testAccCheckJumpCloudAuthPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		return nil
	}
}

// Definindo as provider factories

func testAccJumpCloudAuthPolicyConfig() string {
	return `
resource "jumpcloud_authentication_policy" "test" {
  name        = "test-policy"
  type        = "mfa"
  status      = "active"
  description = "Test Auth Policy"
  settings    = jsonencode({
    mfa = {
      required = true
      methods = ["totp", "push"]
    }
  })
  priority          = 10
  apply_to_all_users = true
}
`
}

// Definindo as provider factories

func TestAccJumpCloudAuthPolicy_basic(t *testing.T) {
	t.Skip("Skipping until CI environment is set up")

	resourceName := "jumpcloud_authentication_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { jctest.TestAccPreCheck(t) },
		Providers:    jctest.TestAccProviders,
		CheckDestroy: testAccCheckJumpCloudAuthPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudAuthPolicyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudAuthPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-test-auth-policy"),
					resource.TestCheckResourceAttr(resourceName, "disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "type", "user_portal"),
				),
			},
		},
	})
}

// Definindo as provider factories

func testAccJumpCloudAuthPolicyConfig_basic() string {
	return `
resource "jumpcloud_authentication_policy" "test" {
  name        = "test-policy"
  type        = "mfa"
  status      = "active"
  description = "Test Auth Policy"
  settings    = jsonencode({
    mfa = {
      required = true
      methods = ["totp", "push"]
    }
  })
  priority          = 10
  apply_to_all_users = true
}
`
}
