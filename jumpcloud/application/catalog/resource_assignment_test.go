package app_catalog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccResourceAssignment_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAssignmentConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("jumpcloud_appcatalog_assignment.test", "id"),
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_assignment.test", "application_id", "test-app-id"),
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_assignment.test", "user_id", "test-user-id"),
				),
			},
		},
	})
}

func TestAccResourceAssignment_group(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAssignmentConfig_group(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("jumpcloud_appcatalog_assignment.test_group", "id"),
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_assignment.test_group", "application_id", "test-app-id"),
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_assignment.test_group", "group_id", "test-group-id"),
				),
			},
		},
	})
}

func TestAccResourceAssignment_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAssignmentConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_assignment.test", "user_id", "test-user-id"),
				),
			},
			{
				Config: testAccResourceAssignmentConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_assignment.test", "user_id", "updated-user-id"),
				),
			},
		},
	})
}

func TestAccResourceAssignment_configuration(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAssignmentConfig_withConfiguration(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_assignment.test_config", "configuration.%", "2"),
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_assignment.test_config", "configuration.username", "testuser"),
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_assignment.test_config", "configuration.role", "admin"),
				),
			},
		},
	})
}

func testAccResourceAssignmentConfig_basic() string {
	return `
resource "jumpcloud_appcatalog_assignment" "test" {
  application_id = "test-app-id"
  user_id        = "test-user-id"
}
`
}

func testAccResourceAssignmentConfig_group() string {
	return `
resource "jumpcloud_appcatalog_assignment" "test_group" {
  application_id = "test-app-id"
  group_id       = "test-group-id"
}
`
}

func testAccResourceAssignmentConfig_updated() string {
	return `
resource "jumpcloud_appcatalog_assignment" "test" {
  application_id = "test-app-id"
  user_id        = "updated-user-id"
}
`
}

func testAccResourceAssignmentConfig_withConfiguration() string {
	return `
resource "jumpcloud_appcatalog_assignment" "test_config" {
  application_id = "test-app-id"
  user_id        = "test-user-id"
  configuration = {
    username = "testuser"
    role     = "admin"
  }
}
`
}
