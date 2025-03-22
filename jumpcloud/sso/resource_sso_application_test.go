package sso

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/testing"
)

func TestAccResourceSSOApplication_basic(t *testing.T) {
	rName := fmt.Sprintf("terraform-test-%s", acctest.RandString(8))
	resourceName := "jumpcloud_sso_application.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSSOApplicationConfig_saml(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "saml"),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
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

func TestAccResourceSSOApplication_update(t *testing.T) {
	rName := fmt.Sprintf("terraform-test-%s", acctest.RandString(8))
	resourceName := "jumpcloud_sso_application.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { jctest.TestAccPreCheck(t) },
		Providers: jctest.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSSOApplicationConfig_saml(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Test SSO application"),
				),
			},
			{
				Config: testAccResourceSSOApplicationConfig_saml_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated SSO application"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "Updated Display Name"),
				),
			},
		},
	})
}

func testAccResourceSSOApplicationConfig_saml(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_sso_application" "test" {
  name        = "%s"
  description = "Test SSO application"
  type        = "saml"
  active      = true
  
  saml {
    entity_id              = "https://example.com/saml"
    assertion_consumer_url = "https://example.com/saml/acs"
    name_id_format         = "email"
    saml_signing_algorithm = "sha256"
    sign_assertion         = true
    sign_response          = true
    
    attribute_statements {
      name        = "email"
      name_format = "unspecified"
      value       = "{{email}}"
    }
  }
}
`, name)
}

func testAccResourceSSOApplicationConfig_saml_updated(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_sso_application" "test" {
  name         = "%s"
  display_name = "Updated Display Name"
  description  = "Updated SSO application"
  type         = "saml"
  active       = true
  
  saml {
    entity_id              = "https://example.com/saml"
    assertion_consumer_url = "https://example.com/saml/acs"
    name_id_format         = "email"
    saml_signing_algorithm = "sha256"
    sign_assertion         = true
    sign_response          = true
    
    attribute_statements {
      name        = "email"
      name_format = "unspecified"
      value       = "{{email}}"
    }
    
    attribute_statements {
      name        = "displayName"
      name_format = "unspecified"
      value       = "{{displayName}}"
    }
  }
}
`, name)
}
