package notifications_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudNotificationChannel_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_notification_channel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudNotificationChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudNotificationChannelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudNotificationChannelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "email"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Configuration is marked sensitive, so it won't be returned during import
				ImportStateVerifyIgnore: []string{"configuration"},
			},
		},
	})
}

func TestAccJumpCloudNotificationChannel_update(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := rName + "-updated"
	resourceName := "jumpcloud_notification_channel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudNotificationChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudNotificationChannelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudNotificationChannelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccJumpCloudNotificationChannelConfig_updated(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudNotificationChannelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "recipients.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "alert_severity.*", "critical"),
					resource.TestCheckTypeSetElemAttr(resourceName, "alert_severity.*", "high"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudNotificationChannelExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		client := commonTesting.TestAccProviders["jumpcloud"].Meta().(interface {
			DoRequest(method, path string, body []byte) ([]byte, error)
		})

		_, err := client.DoRequest("GET", fmt.Sprintf("/api/v2/notification-channels/%s", rs.Primary.ID), nil)
		if err != nil {
			return fmt.Errorf("error fetching notification channel with ID %s: %s", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckJumpCloudNotificationChannelDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_notification_channel" {
			continue
		}

		client := commonTesting.TestAccProviders["jumpcloud"].Meta().(interface {
			DoRequest(method, path string, body []byte) ([]byte, error)
		})

		_, err := client.DoRequest("GET", fmt.Sprintf("/api/v2/notification-channels/%s", rs.Primary.ID), nil)

		// The request should return an error if the notification channel is destroyed
		if err == nil {
			return fmt.Errorf("JumpCloud Notification Channel %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccJumpCloudNotificationChannelConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_notification_channel" "test" {
  name          = %q
  type          = "email"
  enabled       = true
  configuration = jsonencode({
    recipients = ["admin@example.com"]
  })
}
`, rName)
}

func testAccJumpCloudNotificationChannelConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_notification_channel" "test" {
  name          = %q
  type          = "email"
  enabled       = false
  configuration = jsonencode({
    recipients = ["admin@example.com", "ops@example.com"]
  })
  recipients     = ["admin@example.com", "ops@example.com"]
  alert_severity = ["critical", "high"]
  throttling     = jsonencode({
    limit       = 10,
    timeWindow  = 3600,
    cooldown    = 300
  })
}
`, rName)
}
