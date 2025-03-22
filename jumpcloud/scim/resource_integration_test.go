package scim

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccJumpCloudScimIntegration_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "jumpcloud_scim_integration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudScimIntegrationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-test-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "type", "saas"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "sync_schedule", "manual"),
					resource.TestCheckResourceAttrSet(resourceName, "server_id"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"settings"},
			},
		},
	})
}

func TestAccJumpCloudScimIntegration_update(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "jumpcloud_scim_integration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudScimIntegrationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-test-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "sync_schedule", "manual"),
				),
			},
			{
				Config: testAccJumpCloudScimIntegrationConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-test-updated-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "sync_schedule", "daily"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
		},
	})
}

func testAccJumpCloudScimIntegrationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_scim_server" "test" {
  name        = "tf-test-server-%s"
  type        = "generic"
  auth_type   = "token"
  auth_config = jsonencode({
    token = "test-token"
  })
  enabled     = true
}

resource "jumpcloud_scim_integration" "test" {
  name         = "tf-test-%s"
  type         = "saas"
  server_id    = jumpcloud_scim_server.test.id
  enabled      = true
  sync_schedule = "manual"
  settings     = jsonencode({
    service_provider = "test-provider"
    connection_type  = "standard"
  })
}
`, rName, rName)
}

func testAccJumpCloudScimIntegrationConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_scim_server" "test" {
  name        = "tf-test-server-%s"
  type        = "generic"
  auth_type   = "token"
  auth_config = jsonencode({
    token = "test-token"
  })
  enabled     = true
}

resource "jumpcloud_scim_integration" "test" {
  name         = "tf-test-updated-%s"
  description  = "Updated description"
  type         = "saas"
  server_id    = jumpcloud_scim_server.test.id
  enabled      = false
  sync_schedule = "daily"
  settings     = jsonencode({
    service_provider = "updated-provider"
    connection_type  = "enhanced"
    advanced_settings = {
      retry_attempts = 3
      timeout        = 60
    }
  })
}
`, rName, rName)
}
