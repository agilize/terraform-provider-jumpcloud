package commands_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccDataSourceJumpCloudCommand_byName(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "data.jumpcloud_command.test_by_name"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceJumpCloudCommandConfig_byName(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						resourceName, "id",
						"jumpcloud_command.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "command", "echo 'Hello from Terraform'"),
					resource.TestCheckResourceAttr(resourceName, "command_type", "linux"),
					resource.TestCheckResourceAttr(resourceName, "user", "root"),
					resource.TestCheckResourceAttr(resourceName, "sudo", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceJumpCloudCommand_byID(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "data.jumpcloud_command.test_by_id"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceJumpCloudCommandConfig_byID(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						resourceName, "id",
						"jumpcloud_command.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "command", "echo 'Hello from Terraform'"),
					resource.TestCheckResourceAttr(resourceName, "command_type", "linux"),
					resource.TestCheckResourceAttr(resourceName, "user", "root"),
					resource.TestCheckResourceAttr(resourceName, "sudo", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceJumpCloudCommand_withAssociations(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "data.jumpcloud_command.test_with_assoc"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceJumpCloudCommandConfig_withAssociations(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						resourceName, "id",
						"jumpcloud_command.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "systems.#"),
					resource.TestCheckResourceAttrSet(resourceName, "system_groups.#"),
				),
			},
		},
	})
}

func testAccDataSourceJumpCloudCommandConfig_byName(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_command" "test" {
  name         = %[1]q
  command      = "echo 'Hello from Terraform'"
  command_type = "linux"
  user         = "root"
  sudo         = true
}

data "jumpcloud_command" "test_by_name" {
  name = jumpcloud_command.test.name
}
`, rName)
}

func testAccDataSourceJumpCloudCommandConfig_byID(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_command" "test" {
  name         = %[1]q
  command      = "echo 'Hello from Terraform'"
  command_type = "linux"
  user         = "root"
  sudo         = true
}

data "jumpcloud_command" "test_by_id" {
  id = jumpcloud_command.test.id
}
`, rName)
}

func testAccDataSourceJumpCloudCommandConfig_withAssociations(rName string) string {
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

resource "jumpcloud_system_group" "test" {
  name = %[1]q
}

resource "jumpcloud_command_association" "system_test" {
  command_id = jumpcloud_command.test.id
  system_id  = jumpcloud_system.test.id
}

resource "jumpcloud_command_association" "group_test" {
  command_id = jumpcloud_command.test.id
  group_id   = jumpcloud_system_group.test.id
}

data "jumpcloud_command" "test_with_assoc" {
  id = jumpcloud_command.test.id
  depends_on = [
    jumpcloud_command_association.system_test,
    jumpcloud_command_association.group_test
  ]
}
`, rName)
}
