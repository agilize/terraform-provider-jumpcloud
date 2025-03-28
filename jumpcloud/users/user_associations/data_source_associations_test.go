package user_associations_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudDataSourceUserAssociations_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceUserAssociationsConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_user_associations.all", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_user_associations.all", "associations.#"),
				),
			},
		},
	})
}

func TestAccJumpCloudDataSourceUserAssociations_filtered(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceUserAssociationsConfig_filtered(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_user_associations.filtered", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_user_associations.filtered", "associations.#"),
				),
			},
		},
	})
}

func TestAccJumpCloudDataSourceUserAssociations_withUser(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceUserAssociationsConfig_withUser(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_user_associations.with_user", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_user_associations.with_user", "associations.#"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceUserAssociationsConfig_basic() string {
	return `
data "jumpcloud_user_associations" "all" {}
`
}

func testAccJumpCloudDataSourceUserAssociationsConfig_filtered() string {
	return `
data "jumpcloud_user_associations" "filtered" {
  filter {
    field = "type"
    operator = "eq"
    value = "group"
  }
  sort {
    field = "name"
    direction = "asc"
  }
}

output "association_count" {
  value = data.jumpcloud_user_associations.filtered.associations.#
}
`
}

func testAccJumpCloudDataSourceUserAssociationsConfig_withUser() string {
	return `
resource "jumpcloud_user" "test" {
  username = "test-user"
  email = "test@example.com"
  firstname = "Test"
  lastname = "User"
}

resource "jumpcloud_user_association" "test" {
  user_id = jumpcloud_user.test.id
  type = "group"
  target_id = "test-group-id"
}

data "jumpcloud_user_associations" "with_user" {
  filter {
    field = "user_id"
    operator = "eq"
    value = jumpcloud_user.test.id
  }
}

output "association_count" {
  value = data.jumpcloud_user_associations.with_user.associations.#
}
`
}
