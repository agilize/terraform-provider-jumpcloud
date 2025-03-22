package authentication

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDataSourcePolicies(t *testing.T) {
	r := DataSourcePolicies()
	if r == nil {
		t.Fatal("Expected non-nil resource")
	}
	if r.Schema["filter"] == nil {
		t.Fatal("Expected non-nil filter schema")
	}
	if r.Schema["sort"] == nil {
		t.Fatal("Expected non-nil sort schema")
	}
	if r.Schema["limit"] == nil {
		t.Fatal("Expected non-nil limit schema")
	}
	if r.Schema["auth_policies"] == nil {
		t.Fatal("Expected non-nil auth_policies schema")
	}
	if r.ReadContext == nil {
		t.Fatal("Expected non-nil ReadContext")
	}
}

// Uncomment acceptance tests now that the authentication resources are enabled
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
