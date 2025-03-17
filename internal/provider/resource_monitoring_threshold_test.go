package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMonitoringThreshold_basic(t *testing.T) {
	var thresholdID string

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_monitoring_threshold.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMonitoringThresholdDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitoringThresholdConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitoringThresholdExists(resourceName, &thresholdID),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Basic threshold"),
					resource.TestCheckResourceAttr(resourceName, "type", "cpu"),
					resource.TestCheckResourceAttr(resourceName, "threshold", "90"),
					resource.TestCheckResourceAttr(resourceName, "operator", "gt"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
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

func TestAccResourceMonitoringThreshold_update(t *testing.T) {
	var thresholdID string

	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := rName + "-updated"
	resourceName := "jumpcloud_monitoring_threshold.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMonitoringThresholdDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitoringThresholdConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitoringThresholdExists(resourceName, &thresholdID),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "threshold", "90"),
				),
			},
			{
				Config: testAccMonitoringThresholdConfig_updated(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitoringThresholdExists(resourceName, &thresholdID),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated threshold"),
					resource.TestCheckResourceAttr(resourceName, "threshold", "95"),
					resource.TestCheckResourceAttr(resourceName, "duration", "10"),
					resource.TestCheckResourceAttr(resourceName, "system_targets.#", "1"),
				),
			},
		},
	})
}

func testAccCheckMonitoringThresholdExists(resourceName string, thresholdID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		*thresholdID = rs.Primary.ID

		return nil
	}
}

func testAccCheckMonitoringThresholdDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_monitoring_threshold" {
			continue
		}

		// Retrieve the client from the test provider
		client := testAccProvider.Meta().(ClientInterface)

		// Check that the threshold no longer exists
		url := fmt.Sprintf("/api/v2/monitoring/thresholds/%s", rs.Primary.ID)
		_, err := client.DoRequest("GET", url, nil)

		// The request should return an error if the threshold is destroyed
		if err == nil {
			return fmt.Errorf("JumpCloud Monitoring Threshold %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccMonitoringThresholdConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_monitoring_threshold" "test" {
  name        = %q
  description = "Basic threshold"
  type        = "cpu"
  threshold   = 90
  operator    = "gt"
  enabled     = true
}
`, rName)
}

func testAccMonitoringThresholdConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_system" "test" {
  name        = "test-system"
  displayName = "Test System"
  os          = "Linux"
}

resource "jumpcloud_monitoring_threshold" "test" {
  name           = %q
  description    = "Updated threshold"
  type           = "cpu"
  threshold      = 95
  operator       = "gt"
  enabled        = true
  duration       = 10
  system_targets = [jumpcloud_system.test.id]
}
`, rName)
}
