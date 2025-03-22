package appcatalog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/testing"
)

func TestAccResourceAssignment_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAssignmentConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("jumpcloud_appcatalog_assignment.test", "id"),
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_assignment.test", "assignment_type", "required"),
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_assignment.test", "target_type", "user"),
				),
			},
		},
	})
}

func TestAccResourceAssignment_group(t *testing.T) {
	var resourceName = "jumpcloud_appcatalog_assignment.test_group"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAssignmentConfig_group(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "target_type", "group"),
					resource.TestCheckResourceAttr(resourceName, "assignment_type", "optional"),
					resource.TestCheckResourceAttr(resourceName, "install_policy", "manual"),
				),
			},
		},
	})
}

func TestAccResourceAssignment_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAssignmentConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_assignment.test", "assignment_type", "required"),
				),
			},
			{
				Config: testAccResourceAssignmentConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("jumpcloud_appcatalog_assignment.test", "assignment_type", "optional"),
				),
			},
		},
	})
}

func TestAccResourceAssignment_configuration(t *testing.T) {
	var resourceName = "jumpcloud_appcatalog_assignment.test_config"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAssignmentConfig_configuration(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration"),
				),
			},
		},
	})
}

// Config generation functions
func testAccResourceAssignmentConfig_basic() string {
	return `
resource "jumpcloud_appcatalog_application" "test_app" {
  name        = "Test Application for Assignment"
  description = "Application for testing assignments"
  app_type    = "web"
  status      = "active"
  visibility  = "private"
}

resource "jumpcloud_user" "test_user" {
  username  = "testuser"
  email     = "testuser@example.com"
  firstname = "Test"
  lastname  = "User"
}

resource "jumpcloud_appcatalog_assignment" "test" {
  application_id  = jumpcloud_appcatalog_application.test_app.id
  target_type     = "user"
  target_id       = jumpcloud_user.test_user.id
  assignment_type = "required"
}
`
}

func testAccResourceAssignmentConfig_group() string {
	return `
resource "jumpcloud_appcatalog_application" "test_app" {
  name        = "Test Application for Group Assignment"
  description = "Application for testing group assignments"
  app_type    = "web"
  status      = "active"
  visibility  = "private"
}

resource "jumpcloud_user_group" "test_group" {
  name = "Test Group"
}

resource "jumpcloud_appcatalog_assignment" "test_group" {
  application_id  = jumpcloud_appcatalog_application.test_app.id
  target_type     = "group"
  target_id       = jumpcloud_user_group.test_group.id
  assignment_type = "optional"
  install_policy  = "manual"
}
`
}

func testAccResourceAssignmentConfig_updated() string {
	return `
resource "jumpcloud_appcatalog_application" "test_app" {
  name        = "Test Application for Assignment"
  description = "Application for testing assignments"
  app_type    = "web"
  status      = "active"
  visibility  = "private"
}

resource "jumpcloud_user" "test_user" {
  username  = "testuser"
  email     = "testuser@example.com"
  firstname = "Test"
  lastname  = "User"
}

resource "jumpcloud_appcatalog_assignment" "test" {
  application_id  = jumpcloud_appcatalog_application.test_app.id
  target_type     = "user"
  target_id       = jumpcloud_user.test_user.id
  assignment_type = "optional"
  install_policy  = "manual"
}
`
}

func testAccResourceAssignmentConfig_configuration() string {
	return `
resource "jumpcloud_appcatalog_application" "test_app" {
  name        = "Test Application with Configuration"
  description = "Application for testing configuration"
  app_type    = "web"
  status      = "active"
  visibility  = "private"
}

resource "jumpcloud_user" "test_user" {
  username  = "testuser-config"
  email     = "testuser-config@example.com"
  firstname = "Test"
  lastname  = "User"
}

resource "jumpcloud_appcatalog_assignment" "test_config" {
  application_id = jumpcloud_appcatalog_application.test_app.id
  target_type    = "user"
  target_id      = jumpcloud_user.test_user.id
  configuration  = jsonencode({
    key1 = "value1"
    key2 = "value2"
    nested = {
      nestedKey = "nestedValue"
    }
  })
}
`
}
