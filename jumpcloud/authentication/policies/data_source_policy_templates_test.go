package policies

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestDataSourcePolicyTemplatesSchema(t *testing.T) {
	s := DataSourcePolicyTemplates()
	if s == nil {
		t.Fatal("Expected non-nil resource")
	}
	if s.Schema["filter"] == nil {
		t.Fatal("Expected non-nil filter schema")
	}
	if s.Schema["sort"] == nil {
		t.Fatal("Expected non-nil sort schema")
	}
	if s.Schema["limit"] == nil {
		t.Fatal("Expected non-nil limit schema")
	}
	if s.Schema["templates"] == nil {
		t.Fatal("Expected non-nil templates schema")
	}
	if s.ReadContext == nil {
		t.Fatal("Expected non-nil ReadContext")
	}
}

func TestAccDataSourcePolicyTemplates_basic(t *testing.T) {
	t.Skip("Skipping until CI environment is set up")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePolicyTemplatesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_policy_templates.all", "templates.#"),
				),
			},
		},
	})
}

func testAccDataSourcePolicyTemplatesConfig_basic() string {
	return `
data "jumpcloud_policy_templates" "all" {}
`
}
