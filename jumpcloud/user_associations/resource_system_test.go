package user_associations

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccJumpCloudUserSystemAssociation(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	var resourceName = "jumpcloud_user_system_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckJumpCloudUserSystemAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudUserSystemAssociationConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudUserSystemAssociationExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "user_id"),
					resource.TestCheckResourceAttrSet(resourceName, "system_id"),
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

func testAccJumpCloudUserSystemAssociationConfig() string {
	return `
resource "jumpcloud_user" "test" {
  username   = "test-acc-system-association"
  email      = "test-acc-system-association@example.com"
  firstname  = "Test"
  lastname   = "User"
}

resource "jumpcloud_system" "test" {
  displayname = "test-acc-system-association"
  os          = "Linux"
}

resource "jumpcloud_user_system_association" "test" {
  user_id   = jumpcloud_user.test.id
  system_id = jumpcloud_system.test.id
}
`
}

func testAccCheckJumpCloudUserSystemAssociationExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		// Add implementation to check if the association actually exists in JumpCloud
		// This will depend on how you've structured your test setup

		return nil
	}
}

func testAccCheckJumpCloudUserSystemAssociationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_user_system_association" {
			continue
		}

		// Add implementation to check if the association has been destroyed in JumpCloud
		// This will depend on how you've structured your test setup

		return nil
	}

	return nil
}
