package policies

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// Definindo as provider factories

func TestResourcePolicyBinding(t *testing.T) {
	r := ResourcePolicyBinding()
	// Use standard Go testing instead of assert
	if r == nil {
		t.Fatal("Expected non-nil resource")
	}
	if r.Schema["policy_id"] == nil {
		t.Fatal("Expected non-nil policy_id schema")
	}
	if r.Schema["target_id"] == nil {
		t.Fatal("Expected non-nil target_id schema")
	}
	if r.Schema["target_type"] == nil {
		t.Fatal("Expected non-nil target_type schema")
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
// Definindo as provider factories

func TestAccPolicyBinding(t *testing.T) {
	t.Skip("Skipping until CI environment is set up")

	resourceName := "jumpcloud_authentication_policy_binding.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { commonTesting.TestAccPreCheck(t) },
		Providers:    commonTesting.TestAccProviders,
		CheckDestroy: testAccCheckJumpCloudAuthPolicyBindingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudAuthPolicyBindingConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudAuthPolicyBindingExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "policy_id"),
					resource.TestCheckResourceAttrSet(resourceName, "target_id"),
					resource.TestCheckResourceAttr(resourceName, "target_type", "user_group"),
				),
			},
		},
	})
}

// Definindo as provider factories

func testAccCheckJumpCloudAuthPolicyBindingDestroy(s *terraform.State) error {
	// Implementation depends on the test setup
	// This is a placeholder for the actual implementation
	return nil
}

// Definindo as provider factories

func testAccCheckJumpCloudAuthPolicyBindingExists(n string) resource.TestCheckFunc {
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

func testAccJumpCloudAuthPolicyBindingConfig() string {
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
  apply_to_all_users = false
}

resource "jumpcloud_user_group" "test" {
  name = "test-user-group"
}

resource "jumpcloud_authentication_policy_binding" "test" {
  policy_id   = jumpcloud_authentication_policy.test.id
  target_id   = jumpcloud_user_group.test.id
  target_type = "user_group"
  priority    = 10
}
`
}
