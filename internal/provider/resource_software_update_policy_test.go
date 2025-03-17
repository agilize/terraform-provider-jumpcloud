package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceSoftwareUpdatePolicy_basic(t *testing.T) {
	var policyID string

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_software_update_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftwareUpdatePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSoftwareUpdatePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftwareUpdatePolicyExists(resourceName, &policyID),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "os_family", "linux"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "all_packages", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_approve", "false"),
					resource.TestMatchResourceAttr(resourceName, "created", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
					resource.TestMatchResourceAttr(resourceName, "updated", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
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

func TestAccResourceSoftwareUpdatePolicy_update(t *testing.T) {
	var policyID string

	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := rName + "-updated"
	resourceName := "jumpcloud_software_update_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftwareUpdatePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSoftwareUpdatePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftwareUpdatePolicyExists(resourceName, &policyID),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Basic software update policy"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccSoftwareUpdatePolicyConfig_updated(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftwareUpdatePolicyExists(resourceName, &policyID),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated software update policy"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_approve", "true"),
				),
			},
		},
	})
}

func testAccCheckSoftwareUpdatePolicyExists(resourceName string, policyID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		*policyID = rs.Primary.ID

		return nil
	}
}

func testAccCheckSoftwareUpdatePolicyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_software_update_policy" {
			continue
		}

		// Retrieve the client from the test provider
		client := testAccProvider.Meta().(ClientInterface)

		// Check that the policy no longer exists
		url := fmt.Sprintf("/api/v2/software/update-policies/%s", rs.Primary.ID)
		_, err := client.DoRequest("GET", url, nil)

		// The request should return a 404 Not Found if the policy is destroyed
		if err == nil {
			return fmt.Errorf("JumpCloud software update policy %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccSoftwareUpdatePolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_software_update_policy" "test" {
  name        = %q
  description = "Basic software update policy"
  os_family   = "linux"
  enabled     = true
  
  schedule = jsonencode({
    type   = "daily"
    hour   = 3
    minute = 0
  })
  
  all_packages = true
  auto_approve = false
}
`, rName)
}

func testAccSoftwareUpdatePolicyConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_software_update_policy" "test" {
  name        = %q
  description = "Updated software update policy"
  os_family   = "linux"
  enabled     = false
  
  schedule = jsonencode({
    type   = "weekly"
    dayOfWeek = "sunday"
    hour   = 2
    minute = 0
  })
  
  all_packages = true
  auto_approve = true
}
`, rName)
}
