package sso

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccDataSourceSSOApplication(t *testing.T) {
	rName := fmt.Sprintf("terraform-test-%s", acctest.RandString(8))
	resourceName := "jumpcloud_sso_application.test"
	dataSourceName := "data.jumpcloud_sso_application.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { jctest.TestAccPreCheck(t) },
		ProviderFactories: jctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSSOApplicationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					// Check the resource
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "saml"),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),

					// Check the data source
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "type", resourceName, "type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "active", resourceName, "active"),
				),
			},
		},
	})
}

func testAccDataSourceSSOApplicationConfig(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_sso_application" "test" {
  name = "%s"
  type = "saml"
  active = true
  display_name = "%s"
  description = "Test SSO Application"
  sso_url = "https://test.example.com/saml"
  metadata_url = "https://test.example.com/metadata"
  metadata = <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" xmlns:ds="http://www.w3.org/2000/09/xmldsig#" entityID="https://test.example.com">
  <md:SPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <md:AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://test.example.com/acs"/>
  </md:SPSSODescriptor>
</md:EntityDescriptor>
EOF
}

data "jumpcloud_sso_application" "test" {
  name = jumpcloud_sso_application.test.name
}
`, name, name)
}
