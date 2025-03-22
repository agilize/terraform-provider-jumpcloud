package authentication

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

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

// Uncomment acceptance tests now that the authentication resources are enabled
func TestAccJumpCloudAuthPolicy_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckJumpCloudAuthPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudAuthPolicyConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudAuthPolicyExists("jumpcloud_auth_policy.test"),
					resource.TestCheckResourceAttr("jumpcloud_auth_policy.test", "name", "tf-test-auth-policy"),
					resource.TestCheckResourceAttr("jumpcloud_auth_policy.test", "disabled", "false"),
					resource.TestCheckResourceAttr("jumpcloud_auth_policy.test", "type", "user_portal"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudAuthPolicyDestroy(s *terraform.State) error {
	// Implementation depends on the test setup
	// This is a placeholder for the actual implementation
	return nil
}

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

func testAccJumpCloudAuthPolicyConfig() string {
	return `
resource "jumpcloud_auth_policy" "test" {
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
