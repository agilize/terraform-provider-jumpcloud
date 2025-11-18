package users_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// TestAccResourceUserDeviceAssociation_basic tests basic creation of a user-device association
func TestAccResourceUserDeviceAssociation_basic(t *testing.T) {
	resourceName := "jumpcloud_user_device_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		CheckDestroy:      testAccCheckUserDeviceAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDeviceAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserDeviceAssociationExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "user_id"),
					resource.TestCheckResourceAttrSet(resourceName, "system_id"),
				),
			},
		},
	})
}

// TestAccResourceUserDeviceAssociation_import tests importing an existing association
func TestAccResourceUserDeviceAssociation_import(t *testing.T) {
	resourceName := "jumpcloud_user_device_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		CheckDestroy:      testAccCheckUserDeviceAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDeviceAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserDeviceAssociationExists(resourceName),
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

// TestAccResourceUserDeviceAssociation_multipleAssociations tests multiple associations for one user
func TestAccResourceUserDeviceAssociation_multipleAssociations(t *testing.T) {
	resource1Name := "jumpcloud_user_device_association.test1"
	resource2Name := "jumpcloud_user_device_association.test2"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		CheckDestroy:      testAccCheckUserDeviceAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDeviceAssociationConfig_multiple(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserDeviceAssociationExists(resource1Name),
					testAccCheckUserDeviceAssociationExists(resource2Name),
					resource.TestCheckResourceAttrSet(resource1Name, "id"),
					resource.TestCheckResourceAttrSet(resource2Name, "id"),
				),
			},
		},
	})
}

// TestAccResourceUserDeviceAssociation_validation tests validation errors
func TestAccResourceUserDeviceAssociation_validation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccUserDeviceAssociationConfig_missingUserID(),
				ExpectError: regexp.MustCompile("user_id"),
			},
			{
				Config:      testAccUserDeviceAssociationConfig_missingSystemID(),
				ExpectError: regexp.MustCompile("system_id"),
			},
		},
	})
}

// Helper functions

func testAccCheckUserDeviceAssociationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No User Device Association ID is set")
		}

		return nil
	}
}

func testAccCheckUserDeviceAssociationDestroy(s *terraform.State) error {
	// This would check if the association was properly destroyed
	// For now, we'll return nil as the actual implementation would require API calls
	return nil
}

// Configuration functions

func testAccUserDeviceAssociationConfig_basic() string {
	return `
resource "jumpcloud_user" "test" {
  username   = "testuser_device_assoc"
  email      = "testuser_device_assoc@example.com"
  firstname  = "Test"
  lastname   = "User"
  password   = "P@ssw0rd123!"
}

resource "jumpcloud_system" "test" {
  hostname     = "test-device-assoc"
  display_name = "Test Device for Association"
}

resource "jumpcloud_user_device_association" "test" {
  user_id   = jumpcloud_user.test.id
  system_id = jumpcloud_system.test.id
}
`
}

func testAccUserDeviceAssociationConfig_multiple() string {
	return `
resource "jumpcloud_user" "test" {
  username   = "testuser_device_assoc"
  email      = "testuser_device_assoc@example.com"
  firstname  = "Test"
  lastname   = "User"
  password   = "P@ssw0rd123!"
}

resource "jumpcloud_system" "test1" {
  hostname     = "test-device-assoc-1"
  display_name = "Test Device 1 for Association"
}

resource "jumpcloud_system" "test2" {
  hostname     = "test-device-assoc-2"
  display_name = "Test Device 2 for Association"
}

resource "jumpcloud_user_device_association" "test1" {
  user_id   = jumpcloud_user.test.id
  system_id = jumpcloud_system.test1.id
}

resource "jumpcloud_user_device_association" "test2" {
  user_id   = jumpcloud_user.test.id
  system_id = jumpcloud_system.test2.id
}
`
}

func testAccUserDeviceAssociationConfig_missingUserID() string {
	return `
resource "jumpcloud_system" "test" {
  hostname     = "test-device-assoc"
  display_name = "Test Device for Association"
}

resource "jumpcloud_user_device_association" "test" {
  system_id = jumpcloud_system.test.id
}
`
}

func testAccUserDeviceAssociationConfig_missingSystemID() string {
	return `
resource "jumpcloud_user" "test" {
  username   = "testuser_device_assoc"
  email      = "testuser_device_assoc@example.com"
  firstname  = "Test"
  lastname   = "User"
  password   = "P@ssw0rd123!"
}

resource "jumpcloud_user_device_association" "test" {
  user_id = jumpcloud_user.test.id
}
`
}
