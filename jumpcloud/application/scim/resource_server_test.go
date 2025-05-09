package scim

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// Definindo as provider factories

func TestAccJumpCloudScimServer_basic(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "jumpcloud_scim_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudScimServerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-test-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "type", "generic"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "token"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auth_config"},
			},
		},
	})
}

// Definindo as provider factories

func TestAccJumpCloudScimServer_update(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "jumpcloud_scim_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudScimServerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-test-%s", rName)),
				),
			},
			{
				Config: testAccJumpCloudScimServerConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-test-updated-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
		},
	})
}

// Definindo as provider factories

func testAccJumpCloudScimServerConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_scim_server" "test" {
  name        = "tf-test-%s"
  type        = "generic"
  auth_type   = "token"
  auth_config = jsonencode({
    token = "test-token"
  })
  enabled     = true
}
`, rName)
}

// Definindo as provider factories

func testAccJumpCloudScimServerConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_scim_server" "test" {
  name        = "tf-test-updated-%s"
  description = "Updated description"
  type        = "generic"
  auth_type   = "token"
  enabled     = false
}
`, rName)
}

// Test setup variables and functions
// nolint:unused
var testAccProviders map[string]*schema.Provider

// nolint:unused
var testAccProvider *schema.Provider

// testAccPreCheck is the canonical implementation for SCIM tests
// Definindo as provider factories

// nolint:unused
func testAccPreCheck(t *testing.T) {
	// Add any necessary setup logic here, such as checking for required environment variables
	// This is called before each test runs
}
