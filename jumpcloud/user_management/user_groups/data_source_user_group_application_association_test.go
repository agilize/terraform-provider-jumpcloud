package usergroups_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// TestAccDataSourceUsergroupApplicationAssociation_basic tests basic data source functionality
func TestAccDataSourceUsergroupApplicationAssociation_basic(t *testing.T) {
	dataSourceName := "data.jumpcloud_user_group_application_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUsergroupApplicationAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "user_group_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "application_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "associated"),
				),
			},
		},
	})
}

// TestAccDataSourceUsergroupApplicationAssociation_associated tests checking an associated application
func TestAccDataSourceUsergroupApplicationAssociation_associated(t *testing.T) {
	dataSourceName := "data.jumpcloud_user_group_application_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUsergroupApplicationAssociationConfig_associated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "associated", "true"),
				),
			},
		},
	})
}

// TestAccDataSourceUsergroupApplicationAssociation_notAssociated tests checking a non-associated application
func TestAccDataSourceUsergroupApplicationAssociation_notAssociated(t *testing.T) {
	dataSourceName := "data.jumpcloud_user_group_application_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUsergroupApplicationAssociationConfig_notAssociated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "associated", "false"),
				),
			},
		},
	})
}

// Configuration functions

func testAccDataSourceUsergroupApplicationAssociationConfig_basic() string {
	return `
resource "jumpcloud_user_group" "test" {
  name        = "test-group-app-association-ds"
  description = "Test group for application association data source tests"
}

data "jumpcloud_user_group_application_association" "test" {
  user_group_id  = jumpcloud_user_group.test.id
  application_id = "test-application-id"
}
`
}

func testAccDataSourceUsergroupApplicationAssociationConfig_associated() string {
	return `
resource "jumpcloud_user_group" "test" {
  name        = "test-group-app-association-ds"
  description = "Test group for application association data source tests"
}

resource "jumpcloud_user_group_application_association" "test" {
  user_group_id  = jumpcloud_user_group.test.id
  application_id = "test-application-id"
}

data "jumpcloud_user_group_application_association" "test" {
  user_group_id  = jumpcloud_user_group.test.id
  application_id = jumpcloud_user_group_application_association.test.application_id

  depends_on = [jumpcloud_user_group_application_association.test]
}
`
}

func testAccDataSourceUsergroupApplicationAssociationConfig_notAssociated() string {
	return `
resource "jumpcloud_user_group" "test" {
  name        = "test-group-app-association-ds"
  description = "Test group for application association data source tests"
}

data "jumpcloud_user_group_application_association" "test" {
  user_group_id  = jumpcloud_user_group.test.id
  application_id = "non-existent-application-id"
}
`
}

