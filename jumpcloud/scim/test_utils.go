package scim

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactories = map[string]func() (*schema.Provider, error){
	"jumpcloud": func() (*schema.Provider, error) {
		return jumpcloudTestProvider(), nil
	},
}

// jumpcloudTestProvider returns a JumpCloud provider for testing
func jumpcloudTestProvider() *schema.Provider {
	// Create a mock provider that includes the resources and data sources for SCIM
	provider := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"jumpcloud_scim_server":            ResourceServer(),
			"jumpcloud_scim_attribute_mapping": ResourceAttributeMapping(),
			"jumpcloud_scim_integration":       ResourceIntegration(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"jumpcloud_scim_servers": DataSourceServers(),
			"jumpcloud_scim_schema":  DataSourceSchema(),
		},
	}
	return provider
}

// Note: testAccPreCheck is defined elsewhere in the package.
// This comment explains that the function is used in tests but defined in another file.
