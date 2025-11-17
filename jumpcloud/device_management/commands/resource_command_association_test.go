package commands_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudCommandAssociation_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_command_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudCommandAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudCommandAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudCommandAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(
						resourceName, "command_id",
						"jumpcloud_command.test", "id"),
					resource.TestCheckResourceAttrPair(
						resourceName, "system_id",
						"jumpcloud_system.test", "id"),
				),
			},
		},
	})
}

func TestAccJumpCloudCommandAssociation_group(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_command_association.group_test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudCommandAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudCommandAssociationConfig_group(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudCommandAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(
						resourceName, "command_id",
						"jumpcloud_command.test", "id"),
					resource.TestCheckResourceAttrPair(
						resourceName, "group_id",
						"jumpcloud_system_group.test", "id"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudCommandAssociationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		// When checking with API, we need to extract the IDs from the Terraform resource
		commandID := rs.Primary.Attributes["command_id"]
		systemID := rs.Primary.Attributes["system_id"]
		groupID := rs.Primary.Attributes["group_id"]

		client := commonTesting.TestAccProviders["jumpcloud"].Meta().(interface {
			DoRequest(method, path string, body []byte) ([]byte, error)
		})

		// Check based on what type of association this is (system or group)
		if systemID != "" {
			// Check system association
			path := fmt.Sprintf("/api/v2/commands/%s/systems/%s", commandID, systemID)
			_, err := client.DoRequest("GET", path, nil)
			if err != nil {
				return fmt.Errorf("error fetching command-system association with command ID %s and system ID %s: %s",
					commandID, systemID, err)
			}
		} else if groupID != "" {
			// Check group association
			path := fmt.Sprintf("/api/v2/commands/%s/systemgroups/%s", commandID, groupID)
			_, err := client.DoRequest("GET", path, nil)
			if err != nil {
				return fmt.Errorf("error fetching command-group association with command ID %s and group ID %s: %s",
					commandID, groupID, err)
			}
		} else {
			return fmt.Errorf("neither system_id nor group_id defined in resource %s", resourceName)
		}

		return nil
	}
}

func testAccCheckJumpCloudCommandAssociationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_command_association" {
			continue
		}

		// Extract the IDs from the Terraform resource
		commandID := rs.Primary.Attributes["command_id"]
		systemID := rs.Primary.Attributes["system_id"]
		groupID := rs.Primary.Attributes["group_id"]

		client := commonTesting.TestAccProviders["jumpcloud"].Meta().(interface {
			DoRequest(method, path string, body []byte) ([]byte, error)
		})

		// Check based on what type of association this was (system or group)
		if systemID != "" {
			// Check system association
			path := fmt.Sprintf("/api/v2/commands/%s/systems/%s", commandID, systemID)
			_, err := client.DoRequest("GET", path, nil)

			// The request should return an error if the association is destroyed
			if err == nil {
				return fmt.Errorf("JumpCloud Command-System Association (%s:%s) still exists",
					commandID, systemID)
			}
		} else if groupID != "" {
			// Check group association
			path := fmt.Sprintf("/api/v2/commands/%s/systemgroups/%s", commandID, groupID)
			_, err := client.DoRequest("GET", path, nil)

			// The request should return an error if the association is destroyed
			if err == nil {
				return fmt.Errorf("JumpCloud Command-Group Association (%s:%s) still exists",
					commandID, groupID)
			}
		}
	}

	return nil
}

func testAccJumpCloudCommandAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_command" "test" {
  name         = %[1]q
  command      = "echo 'Hello from Terraform'"
  command_type = "linux"
  user         = "root"
  sudo         = true
}

resource "jumpcloud_system" "test" {
  display_name = %[1]q
  # Additional required properties for the system
}

resource "jumpcloud_command_association" "test" {
  command_id = jumpcloud_command.test.id
  system_id  = jumpcloud_system.test.id
}
`, rName)
}

func testAccJumpCloudCommandAssociationConfig_group(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_command" "test" {
  name         = %[1]q
  command      = "echo 'Hello from Terraform'"
  command_type = "linux"
  user         = "root"
  sudo         = true
}

resource "jumpcloud_system_group" "test" {
  name = %[1]q
}

resource "jumpcloud_command_association" "group_test" {
  command_id = jumpcloud_command.test.id
  group_id   = jumpcloud_system_group.test.id
}
`, rName)
}
