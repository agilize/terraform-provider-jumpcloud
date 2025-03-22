package authentication

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDataSourcePolicyTemplates(t *testing.T) {
	r := DataSourcePolicyTemplates()
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
	if r.Schema["templates"] == nil {
		t.Fatal("Expected non-nil templates schema")
	}
	if r.ReadContext == nil {
		t.Fatal("Expected non-nil ReadContext")
	}
}

// Uncomment acceptance tests now that the authentication resources are enabled
func TestAccJumpCloudAuthPolicyTemplates_basic(t *testing.T) {
	resourceName := "data.jumpcloud_auth_policy_templates.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudAuthPolicyTemplatesConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "templates.#"),
					resource.TestCheckResourceAttrSet(resourceName, "total_count"),
				),
			},
		},
	})
}

func testAccJumpCloudAuthPolicyTemplatesConfig() string {
	return `
data "jumpcloud_auth_policy_templates" "test" {
  limit = 100
}
`
}
