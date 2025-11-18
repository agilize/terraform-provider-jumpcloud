package policies

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// Definindo as provider factories

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

// TestAccAuthPolicies_basic tests the authentication policies data source
// Definindo as provider factories

func TestAccAuthPolicies_basic(t *testing.T) {
	t.Skip("Skipping until CI environment is set up")

	resourceName := "data.jumpcloud_auth_policies.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
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

// Definindo as provider factories

func testAccJumpCloudAuthPoliciesConfig() string {
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

data "jumpcloud_authentication_policies" "test" {
  limit = 100
  depends_on = [jumpcloud_authentication_policy.test]
}
`
}

// Definindo as provider factories

func TestAccDataSourcePolicies(t *testing.T) {
	t.Skip("Skipping until CI environment is set up")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePoliciesConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_authentication_policies.all", "policies.#"),
				),
			},
		},
	})
}

// Definindo as provider factories

func testAccDataSourcePoliciesConfig() string {
	return `
data "jumpcloud_authentication_policies" "all" {}
`
}

// Definindo as provider factories

// nolint:unused
func testAccDataSourceJumpCloudAuthPoliciesConfig() string {
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

data "jumpcloud_auth_policies" "test" {
  limit = 100
  depends_on = [jumpcloud_authentication_policy.test]
}
`
}
