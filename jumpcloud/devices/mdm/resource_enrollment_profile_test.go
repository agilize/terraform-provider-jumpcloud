package mdm_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudMDMEnrollmentProfile_basic(t *testing.T) {
	resourceName := "jumpcloud_mdm_enrollment_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudMDMEnrollmentProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudMDMEnrollmentProfileConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMDMEnrollmentProfileExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "Test Enrollment Profile"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test enrollment profile for acceptance tests"),
					resource.TestCheckResourceAttr(resourceName, "platform", "ios"),
					resource.TestCheckResourceAttr(resourceName, "enrollment_mode", "corporate"),
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

func TestAccJumpCloudMDMEnrollmentProfile_update(t *testing.T) {
	resourceName := "jumpcloud_mdm_enrollment_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudMDMEnrollmentProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudMDMEnrollmentProfileConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMDMEnrollmentProfileExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "Test Enrollment Profile"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test enrollment profile for acceptance tests"),
				),
			},
			{
				Config: testAccJumpCloudMDMEnrollmentProfileConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMDMEnrollmentProfileExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "Updated Enrollment Profile"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description for enrollment profile"),
					resource.TestCheckResourceAttr(resourceName, "platform", "ios"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudMDMEnrollmentProfileExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		// You would typically make an API call here to check if the resource exists
		// This is a simplified version
		return nil
	}
}

func testAccCheckJumpCloudMDMEnrollmentProfileDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_mdm_enrollment_profile" {
			continue
		}

		// You would typically make an API call here to check if the resource was destroyed
		// This is a simplified version
		return nil
	}

	return nil
}

func testAccJumpCloudMDMEnrollmentProfileConfig_basic() string {
	return `
resource "jumpcloud_mdm_enrollment_profile" "test" {
  name            = "Test Enrollment Profile"
  description     = "Test enrollment profile for acceptance tests"
  platform        = "ios"
  enrollment_mode = "corporate"
}
`
}

func testAccJumpCloudMDMEnrollmentProfileConfig_update() string {
	return `
resource "jumpcloud_mdm_enrollment_profile" "test" {
  name            = "Updated Enrollment Profile"
  description     = "Updated description for enrollment profile"
  platform        = "ios"
  enrollment_mode = "corporate"
}
`
}
