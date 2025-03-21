package authentication

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
)

func TestDataSourcePolicies(t *testing.T) {
	r := DataSourcePolicies()
	assert.NotNil(t, r)
	assert.NotNil(t, r.Schema["filter"])
	assert.NotNil(t, r.Schema["sort"])
	assert.NotNil(t, r.Schema["limit"])
	assert.NotNil(t, r.Schema["auth_policies"])
	assert.NotNil(t, r.ReadContext)
}

func TestAccJumpCloudAuthPolicies_basic(t *testing.T) {
	resourceName := "data.jumpcloud_auth_policies.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudAuthPoliciesConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "auth_policies.#"),
					resource.TestCheckResourceAttrSet(resourceName, "total_count"),
				),
			},
		},
	})
}

func testAccJumpCloudAuthPoliciesConfig() string {
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

data "jumpcloud_auth_policies" "test" {
  limit = 100
  depends_on = [jumpcloud_auth_policy.test]
}
`
}
