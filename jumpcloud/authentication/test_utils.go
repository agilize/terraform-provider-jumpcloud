package authentication

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testAccProviders is the map of providers used for acceptance testing
var testAccProviders map[string]*schema.Provider

func init() {
	// Create a simplified provider for testing
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_API_KEY", nil),
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_ORG_ID", nil),
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_API_URL", "https://console.jumpcloud.com/api"),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"jumpcloud_auth_policy":             ResourcePolicy(),
			"jumpcloud_auth_policy_binding":     ResourcePolicyBinding(),
			"jumpcloud_conditional_access_rule": ResourceConditionalAccessRule(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"jumpcloud_auth_policies":         DataSourcePolicies(),
			"jumpcloud_auth_policy_templates": DataSourcePolicyTemplates(),
		},
	}

	testAccProviders = map[string]*schema.Provider{
		"jumpcloud": provider,
	}
}

// testAccPreCheck validates required environment variables are set for acceptance tests
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("JUMPCLOUD_API_KEY"); v == "" {
		t.Fatal("JUMPCLOUD_API_KEY must be set for acceptance tests")
	}
}
