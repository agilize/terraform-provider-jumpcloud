package commands_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudCommandSchedule_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_command_schedule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudCommandScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudCommandScheduleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudCommandScheduleExists(resourceName),
					resource.TestCheckResourceAttrPair(
						resourceName, "command_id",
						"jumpcloud_command.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "trigger", "one-time"),
					resource.TestCheckResourceAttr(resourceName, "launch_type", "immediate"),
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

func TestAccJumpCloudCommandSchedule_update(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_command_schedule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudCommandScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudCommandScheduleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudCommandScheduleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "trigger", "one-time"),
					resource.TestCheckResourceAttr(resourceName, "launch_type", "immediate"),
				),
			},
			{
				Config: testAccJumpCloudCommandScheduleConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudCommandScheduleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "trigger", "recurring"),
					resource.TestCheckResourceAttr(resourceName, "launch_type", "scheduled"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudCommandScheduleExists(resourceName string) resource.TestCheckFunc {
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

		_, err := client.DoRequest("GET", fmt.Sprintf("/api/v2/command/schedules/%s", rs.Primary.ID), nil)
		if err != nil {
			return fmt.Errorf("error fetching command schedule with ID %s: %s", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckJumpCloudCommandScheduleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_command_schedule" {
			continue
		}

		client := commonTesting.TestAccProviders["jumpcloud"].Meta().(interface {
			DoRequest(method, path string, body []byte) ([]byte, error)
		})

		_, err := client.DoRequest("GET", fmt.Sprintf("/api/v2/command/schedules/%s", rs.Primary.ID), nil)

		// The request should return an error if the command schedule is destroyed
		if err == nil {
			return fmt.Errorf("JumpCloud Command Schedule %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccJumpCloudCommandScheduleConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_command" "test" {
  name         = %[1]q
  command      = "echo 'Hello from Terraform'"
  command_type = "linux"
  user         = "root"
  sudo         = true
}

resource "jumpcloud_command_schedule" "test" {
  name       = %[1]q
  command_id = jumpcloud_command.test.id
  schedule   = "* * * * *"
}
`, rName)
}

func testAccJumpCloudCommandScheduleConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_command" "test" {
  name         = %[1]q
  command      = "echo 'Hello from Terraform'"
  command_type = "linux"
  user         = "root"
  sudo         = true
}

resource "jumpcloud_command_schedule" "test" {
  name            = %[1]q
  command_id      = jumpcloud_command.test.id
  schedule        = "*/10 * * * *"
  schedule_repeat = 3
  description     = "Updated schedule description"
}
`, rName)
}
