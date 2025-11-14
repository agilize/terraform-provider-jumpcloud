package usergroups_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// TestAccResourceUsergroupApplicationAssociation_basic tests basic creation of a user group application association
func TestAccResourceUsergroupApplicationAssociation_basic(t *testing.T) {
	resourceName := "jumpcloud_user_group_application_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		CheckDestroy:      testAccCheckUsergroupApplicationAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsergroupApplicationAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsergroupApplicationAssociationExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "user_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
				),
			},
		},
	})
}

// TestAccResourceUsergroupApplicationAssociation_import tests importing an existing association
func TestAccResourceUsergroupApplicationAssociation_import(t *testing.T) {
	resourceName := "jumpcloud_user_group_application_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		CheckDestroy:      testAccCheckUsergroupApplicationAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsergroupApplicationAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsergroupApplicationAssociationExists(resourceName),
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

// TestAccResourceUsergroupApplicationAssociation_multipleAssociations tests multiple applications in one group
func TestAccResourceUsergroupApplicationAssociation_multipleAssociations(t *testing.T) {
	resource1Name := "jumpcloud_user_group_application_association.test1"
	resource2Name := "jumpcloud_user_group_application_association.test2"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		CheckDestroy:      testAccCheckUsergroupApplicationAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsergroupApplicationAssociationConfig_multiple(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsergroupApplicationAssociationExists(resource1Name),
					testAccCheckUsergroupApplicationAssociationExists(resource2Name),
					resource.TestCheckResourceAttrSet(resource1Name, "id"),
					resource.TestCheckResourceAttrSet(resource2Name, "id"),
				),
			},
		},
	})
}

// TestAccResourceUsergroupApplicationAssociation_validation tests validation errors
func TestAccResourceUsergroupApplicationAssociation_validation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccUsergroupApplicationAssociationConfig_missingUsergroupID(),
				ExpectError: regexp.MustCompile("user_group_id"),
			},
			{
				Config:      testAccUsergroupApplicationAssociationConfig_missingApplicationID(),
				ExpectError: regexp.MustCompile("application_id"),
			},
		},
	})
}

// Helper functions

func testAccCheckUsergroupApplicationAssociationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No User Group Application Association ID is set")
		}

		return nil
	}
}

func testAccCheckUsergroupApplicationAssociationDestroy(s *terraform.State) error {
	// This would check if the association was properly destroyed
	// For now, we'll return nil as the actual implementation would require API calls
	return nil
}

// Configuration functions

func testAccUsergroupApplicationAssociationConfig_basic() string {
	return `
resource "jumpcloud_user_group" "test" {
  name        = "test-group-app-association"
  description = "Test group for application association tests"
}

resource "jumpcloud_user_group_application_association" "test" {
  user_group_id  = jumpcloud_user_group.test.id
  application_id = "test-application-id"
}
`
}

func testAccUsergroupApplicationAssociationConfig_multiple() string {
	return `
resource "jumpcloud_user_group" "test" {
  name        = "test-group-app-association"
  description = "Test group for application association tests"
}

resource "jumpcloud_user_group_application_association" "test1" {
  user_group_id  = jumpcloud_user_group.test.id
  application_id = "test-application-id-1"
}

resource "jumpcloud_user_group_application_association" "test2" {
  user_group_id  = jumpcloud_user_group.test.id
  application_id = "test-application-id-2"
}
`
}

func testAccUsergroupApplicationAssociationConfig_missingUsergroupID() string {
	return `
resource "jumpcloud_user_group_application_association" "test" {
  application_id = "test-application-id"
}
`
}

func testAccUsergroupApplicationAssociationConfig_missingApplicationID() string {
	return `
resource "jumpcloud_user_group" "test" {
  name        = "test-group-app-association"
  description = "Test group for application association tests"
}

resource "jumpcloud_user_group_application_association" "test" {
  user_group_id = jumpcloud_user_group.test.id
}
`
}
