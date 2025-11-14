package monitoring_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudMonitoringThreshold_basic(t *testing.T) {
	var thresholdID string

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_monitoring_threshold.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudMonitoringThresholdDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudMonitoringThresholdConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMonitoringThresholdExists(resourceName, &thresholdID),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Basic threshold"),
					resource.TestCheckResourceAttr(resourceName, "metric_type", "cpu"),
					resource.TestCheckResourceAttr(resourceName, "resource_type", "system"),
					resource.TestCheckResourceAttr(resourceName, "threshold", "90"),
					resource.TestCheckResourceAttr(resourceName, "operator", "gt"),
					resource.TestCheckResourceAttr(resourceName, "duration", "300"),
					resource.TestMatchResourceAttr(resourceName, "created", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
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

func TestAccJumpCloudMonitoringThreshold_update(t *testing.T) {
	var thresholdID string

	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := rName + "-updated"
	resourceName := "jumpcloud_monitoring_threshold.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudMonitoringThresholdDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudMonitoringThresholdConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMonitoringThresholdExists(resourceName, &thresholdID),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "threshold", "90"),
				),
			},
			{
				Config: testAccJumpCloudMonitoringThresholdConfig_updated(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMonitoringThresholdExists(resourceName, &thresholdID),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated threshold"),
					resource.TestCheckResourceAttr(resourceName, "threshold", "95"),
					resource.TestCheckResourceAttr(resourceName, "duration", "600"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudMonitoringThresholdExists(resourceName string, thresholdID *string) resource.TestCheckFunc {
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

func testAccCheckJumpCloudMonitoringThresholdDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_monitoring_threshold" {
			continue
		}

		// Retrieve the client from the test provider
		client := commonTesting.TestAccProviders["jumpcloud"].Meta().(interface {
			DoRequest(method, path string, body []byte) ([]byte, error)
		})

		// Check that the threshold no longer exists
		url := fmt.Sprintf("/api/v2/monitoring-thresholds/%s", rs.Primary.ID)
		_, err := client.DoRequest("GET", url, nil)

		// The request should return an error if the threshold is destroyed
		if err == nil {
			return fmt.Errorf("JumpCloud Monitoring Threshold %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccJumpCloudMonitoringThresholdConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_monitoring_threshold" "test" {
  name          = %q
  description   = "Basic threshold"
  metric_type   = "cpu"
  resource_type = "system"
  threshold     = 90
  operator      = "gt"
  duration      = 300
  severity      = "medium"
}
`, rName)
}

func testAccJumpCloudMonitoringThresholdConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_monitoring_threshold" "test" {
  name          = %q
  description   = "Updated threshold"
  metric_type   = "cpu"
  resource_type = "system"
  threshold     = 95
  operator      = "gt"
  duration      = 600
  severity      = "high"
  tags          = ["production", "critical"]
  actions       = jsonencode({
    notification: {
      type: "email",
      recipients: ["admin@example.com"]
    }
  })
}
`, rName)
}
