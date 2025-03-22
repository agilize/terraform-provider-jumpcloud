package sso

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// generateUniqueName generates a unique resource name for testing
func generateUniqueName(prefix string, t *testing.T) string {
	return fmt.Sprintf("%s-%s", prefix, t.Name())
}

// skipIfNotAcceptanceTest skips a test if not running in acceptance test mode
func skipIfNotAcceptanceTest(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping test as TF_ACC is not set")
	}
}

// checkSSOApplicationExists checks if the SSO application exists
func checkSSOApplicationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSO Application ID is set")
		}

		return nil
	}
}

// getSSOApplicationTestConfig returns a test configuration for an SSO application
func getSSOApplicationTestConfig(name, appType string) string {
	switch appType {
	case "saml":
		return fmt.Sprintf(`
resource "jumpcloud_sso_application" "test" {
  name         = "%s"
  display_name = "Terraform Test SAML"
  description  = "Test SAML application created by Terraform"
  type         = "saml"
  sso_url      = "https://example.com/sso"
  active       = true
  
  saml {
    entity_id = "https://example.com/saml/metadata"
    assertion_consumer_url = "https://example.com/saml/acs"
    name_id_format = "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
    sign_assertion = true
    sign_response = true
  }
}`, name)
	case "oidc":
		return fmt.Sprintf(`
resource "jumpcloud_sso_application" "test" {
  name         = "%s"
  display_name = "Terraform Test OIDC"
  description  = "Test OIDC application created by Terraform"
  type         = "oidc"
  sso_url      = "https://example.com/oauth"
  active       = true
  
  oidc {
    redirect_uris  = ["https://example.com/callback"]
    response_types = ["code"]
    grant_types    = ["authorization_code", "refresh_token"]
    scopes         = ["openid", "profile", "email"]
  }
}`, name)
	default:
		return ""
	}
}
