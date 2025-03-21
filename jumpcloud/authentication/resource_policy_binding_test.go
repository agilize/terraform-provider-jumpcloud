package authentication

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestResourcePolicyBinding(t *testing.T) {
	r := ResourcePolicyBinding()
	assert.NotNil(t, r)
	assert.NotNil(t, r.Schema["policy_id"])
	assert.NotNil(t, r.Schema["target_id"])
	assert.NotNil(t, r.Schema["target_type"])
	assert.NotNil(t, r.CreateContext)
	assert.NotNil(t, r.ReadContext)
	assert.NotNil(t, r.UpdateContext)
	assert.NotNil(t, r.DeleteContext)
}

func TestAccJumpCloudAuthPolicyBinding_basic(t *testing.T) {
	resourceName := "jumpcloud_auth_policy_binding.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckJumpCloudAuthPolicyBindingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudAuthPolicyBindingConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudAuthPolicyBindingExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "policy_id"),
					resource.TestCheckResourceAttrSet(resourceName, "target_id"),
					resource.TestCheckResourceAttr(resourceName, "target_type", "user_group"),
					resource.TestCheckResourceAttr(resourceName, "priority", "10"),
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

func testAccCheckJumpCloudAuthPolicyBindingDestroy(s *terraform.State) error {
	// Implementation depends on the test setup
	// This is a placeholder for the actual implementation
	return nil
}

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

func testAccJumpCloudAuthPolicyBindingConfig() string {
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
  apply_to_all_users = false
}

resource "jumpcloud_user_group" "test" {
  name = "test-user-group"
}

resource "jumpcloud_auth_policy_binding" "test" {
  policy_id   = jumpcloud_auth_policy.test.id
  target_id   = jumpcloud_user_group.test.id
  target_type = "user_group"
  priority    = 10
}
`
}
