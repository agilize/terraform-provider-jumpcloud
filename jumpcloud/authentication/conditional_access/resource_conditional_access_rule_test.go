package authentication

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// Definindo as provider factories

func TestResourceConditionalAccessRule(t *testing.T) {
	r := ResourceConditionalAccessRule()
	// Use standard Go testing instead of assert
	if r == nil {
		t.Fatal("Expected non-nil resource")
	}
	if r.Schema["name"] == nil {
		t.Fatal("Expected non-nil name schema")
	}
	if r.Schema["policy_id"] == nil {
		t.Fatal("Expected non-nil policy_id schema")
	}
	if r.Schema["conditions"] == nil {
		t.Fatal("Expected non-nil conditions schema")
	}
	if r.Schema["action"] == nil {
		t.Fatal("Expected non-nil action schema")
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

func TestAccConditionalAccessRule(t *testing.T) {
	t.Skip("Skipping until CI environment is set up")

	resourceName := "jumpcloud_authentication_conditional_access_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { commonTesting.TestAccPreCheck(t) },
		Providers:    commonTesting.TestAccProviders,
		CheckDestroy: testAccCheckJumpCloudConditionalAccessRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudConditionalAccessRuleConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudConditionalAccessRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-test-rule"),
					resource.TestCheckResourceAttr(resourceName, "action", "deny"),
				),
			},
		},
	})
}

// Definindo as provider factories

func testAccCheckJumpCloudConditionalAccessRuleDestroy(s *terraform.State) error {
	// Implementation depends on the test setup
	// This is a placeholder for the actual implementation
	return nil
}

// Definindo as provider factories

func testAccCheckJumpCloudConditionalAccessRuleExists(n string) resource.TestCheckFunc {
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

func testAccJumpCloudConditionalAccessRuleConfig() string {
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

resource "jumpcloud_authentication_conditional_access_rule" "test" {
  name        = "test-rule"
  status      = "active"
  description = "Test Conditional Access Rule"
  policy_id   = jumpcloud_authentication_policy.test.id
  conditions  = jsonencode({
    network = {
      include = ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]
    }
  })
  action   = "allow"
  priority = 10
}
`
}
