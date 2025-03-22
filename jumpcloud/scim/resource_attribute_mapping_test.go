package scim

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccJumpCloudScimAttributeMapping_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "jumpcloud_scim_attribute_mapping.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudScimAttributeMappingConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-test-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "direction", "inbound"),
					resource.TestCheckResourceAttr(resourceName, "auto_generate", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "server_id"),
					resource.TestCheckResourceAttrSet(resourceName, "schema_id"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
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

func TestAccJumpCloudScimAttributeMapping_update(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "jumpcloud_scim_attribute_mapping.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudScimAttributeMappingConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-test-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "direction", "inbound"),
					resource.TestCheckResourceAttr(resourceName, "object_class", ""),
				),
			},
			{
				Config: testAccJumpCloudScimAttributeMappingConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-test-updated-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "direction", "bidirectional"),
					resource.TestCheckResourceAttr(resourceName, "object_class", "User"),
				),
			},
		},
	})
}

func testAccJumpCloudScimAttributeMappingConfig_basic(rName string) string {
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

resource "jumpcloud_scim_attribute_mapping" "test" {
  name      = "tf-test-%s"
  server_id = jumpcloud_scim_server.test.id
  schema_id = "urn:ietf:params:scim:schemas:core:2.0:User"
  direction = "inbound"
  mappings  = [
    {
      source_path = "name.givenName"
      target_path = "firstName"
      required    = true
      multivalued = false
    },
    {
      source_path = "name.familyName"
      target_path = "lastName"
      required    = true
      multivalued = false
    }
  ]
}
`, rName, rName)
}

func testAccJumpCloudScimAttributeMappingConfig_update(rName string) string {
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

resource "jumpcloud_scim_attribute_mapping" "test" {
  name        = "tf-test-updated-%s"
  description = "Updated description"
  server_id   = jumpcloud_scim_server.test.id
  schema_id   = "urn:ietf:params:scim:schemas:core:2.0:User"
  direction   = "bidirectional"
  object_class = "User"
  mappings    = [
    {
      source_path = "name.givenName"
      target_path = "firstName"
      required    = true
      multivalued = false
      transform   = "toLowerCase"
    },
    {
      source_path = "name.familyName"
      target_path = "lastName"
      required    = true
      multivalued = false
    },
    {
      source_path = "emails[type eq \"work\"].value"
      target_path = "email"
      required    = false
      multivalued = true
    }
  ]
}
`, rName, rName)
}
