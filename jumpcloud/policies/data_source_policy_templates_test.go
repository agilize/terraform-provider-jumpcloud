package policies

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
)

func TestDataSourcePolicyTemplates(t *testing.T) {
	r := DataSourcePolicyTemplates()
	assert.NotNil(t, r)
	assert.NotNil(t, r.Schema["filter"])
	assert.NotNil(t, r.Schema["sort"])
	assert.NotNil(t, r.Schema["limit"])
	assert.NotNil(t, r.Schema["templates"])
	assert.NotNil(t, r.ReadContext)
}

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
