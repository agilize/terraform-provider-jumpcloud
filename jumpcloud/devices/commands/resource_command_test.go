package commands_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudCommand_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_command.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudCommandDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudCommandConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudCommandExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "command", "echo 'Hello from Terraform'"),
					resource.TestCheckResourceAttr(resourceName, "command_type", "linux"),
					resource.TestCheckResourceAttr(resourceName, "sudo", "true"),
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

func TestAccJumpCloudCommand_update(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := rName + "-updated"
	resourceName := "jumpcloud_command.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudCommandDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudCommandConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudCommandExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "command", "echo 'Hello from Terraform'"),
					resource.TestCheckResourceAttr(resourceName, "timeout", "120"),
				),
			},
			{
				Config: testAccJumpCloudCommandConfig_updated(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudCommandExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "command", "echo 'Hello updated from Terraform'"),
					resource.TestCheckResourceAttr(resourceName, "timeout", "240"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudCommandExists(resourceName string) resource.TestCheckFunc {
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

		_, err := client.DoRequest("GET", fmt.Sprintf("/api/commands/%s", rs.Primary.ID), nil)
		if err != nil {
			return fmt.Errorf("error fetching command with ID %s: %s", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckJumpCloudCommandDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_command" {
			continue
		}

		client := commonTesting.TestAccProviders["jumpcloud"].Meta().(interface {
			DoRequest(method, path string, body []byte) ([]byte, error)
		})

		_, err := client.DoRequest("GET", fmt.Sprintf("/api/commands/%s", rs.Primary.ID), nil)

		// The request should return an error if the command is destroyed
		if err == nil {
			return fmt.Errorf("JumpCloud Command %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccJumpCloudCommandConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_command" "test" {
  name         = %q
  command      = "echo 'Hello from Terraform'"
  command_type = "linux"
  user         = "root"
  sudo         = true
  timeout      = 120
}
`, rName)
}

func testAccJumpCloudCommandConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_command" "test" {
  name         = %q
  command      = "echo 'Hello updated from Terraform'"
  command_type = "linux"
  user         = "root"
  sudo         = true
  timeout      = 240
  description  = "Updated test command"
}
`, rName)
}
