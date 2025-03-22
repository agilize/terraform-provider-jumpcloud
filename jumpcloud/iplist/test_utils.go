package iplist

import (
	"testing"

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

// jumpcloudTestProvider returns a mock provider for testing
func jumpcloudTestProvider() *schema.Provider {
	provider := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"jumpcloud_ip_list":            ResourceList(),
			"jumpcloud_ip_list_assignment": ResourceListAssignment(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"jumpcloud_ip_lists":     DataSourceLists(),
			"jumpcloud_ip_locations": DataSourceLocations(),
		},
	}
	return provider
}

// testAccPreCheck is a helper function for acceptance tests
func testAccPreCheck(t *testing.T) {
	// Add any setup logic here
}
