package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceScimServer_basic(t *testing.T) {
	var serverID string

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_scim_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScimServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScimServerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScimServerExists(resourceName, &serverID),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Basic SCIM server"),
					resource.TestCheckResourceAttr(resourceName, "type", "azure"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "basic"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestMatchResourceAttr(resourceName, "created", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
					resource.TestMatchResourceAttr(resourceName, "updated", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"}, // Token não é retornado nas API calls GET
			},
		},
	})
}

func TestAccResourceScimServer_update(t *testing.T) {
	var serverID string

	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := rName + "-updated"
	resourceName := "jumpcloud_scim_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScimServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScimServerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScimServerExists(resourceName, &serverID),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Basic SCIM server"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccScimServerConfig_updated(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScimServerExists(resourceName, &serverID),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated SCIM server"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
		},
	})
}

func testAccCheckScimServerExists(resourceName string, serverID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		*serverID = rs.Primary.ID

		return nil
	}
}

func testAccCheckScimServerDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_scim_server" {
			continue
		}

		// Retrieve the client from the test provider
		client := testAccProvider.Meta().(ClientInterface)

		// Check that the SCIM server no longer exists
		url := fmt.Sprintf("/api/v2/scim/servers/%s", rs.Primary.ID)
		_, err := client.DoRequest("GET", url, nil)

		// The request should return an error if the server is destroyed
		if err == nil {
			return fmt.Errorf("JumpCloud SCIM Server %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScimServerConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_scim_server" "test" {
  name        = %q
  description = "Basic SCIM server"
  type        = "azure"
  auth_type   = "basic"
  enabled     = true
}
`, rName)
}

func testAccScimServerConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_scim_server" "test" {
  name        = %q
  description = "Updated SCIM server"
  type        = "azure"
  auth_type   = "basic"
  enabled     = false
}
`, rName)
}
