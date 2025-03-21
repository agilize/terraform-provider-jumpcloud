package jumpcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/appcatalog"
	"registry.terraform.io/agilize/jumpcloud/pkg/apiclient"
)

// New returns a provider plugin instance
func New() *schema.Provider {
	return Provider()
}

// Provider returns a schema.Provider for JumpCloud.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_API_KEY", nil),
				Description: "API key for JumpCloud operations.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_ORG_ID", nil),
				Description: "Organization ID for JumpCloud multi-tenant environments.",
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_API_URL", "https://console.jumpcloud.com/api"),
				Description: "JumpCloud API URL.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			// App Catalog - Resources
			"jumpcloud_appcatalog_application": appcatalog.ResourceAppCatalogApplication(),
			"jumpcloud_appcatalog_assignment":  appcatalog.ResourceAssignment(),
			"jumpcloud_appcatalog_category":    appcatalog.ResourceCategory(),

			// Legacy resource names - will be deprecated in future versions
			"jumpcloud_app_catalog_application": appcatalog.ResourceAppCatalogApplication(),
			"jumpcloud_app_catalog_assignment":  appcatalog.ResourceAssignment(),
			"jumpcloud_app_catalog_category":    appcatalog.ResourceCategory(),

			// TODO: Move the remaining resources to their appropriate domain packages
			// and update the imports here
		},
		DataSourcesMap: map[string]*schema.Resource{
			// App Catalog - Data Sources
			"jumpcloud_appcatalog_application":  appcatalog.DataSourceApplication(),
			"jumpcloud_appcatalog_applications": appcatalog.DataSourceAppCatalogApplications(),

			// Legacy data source names - will be deprecated in future versions
			"jumpcloud_app_catalog_applications": appcatalog.DataSourceAppCatalogApplications(),

			// TODO: Move the remaining data sources to their appropriate domain packages
			// and update the imports here
		},
		ConfigureContextFunc: providerConfigure,
	}
}

// providerConfigure configures the provider with authentication details
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	tflog.Info(ctx, "Configuring JumpCloud client")

	apiKey := d.Get("api_key").(string)
	orgID := d.Get("org_id").(string)
	apiURL := d.Get("api_url").(string)

	config := &apiclient.Config{
		APIKey: apiKey,
		OrgID:  orgID,
		APIURL: apiURL,
	}

	client := apiclient.NewClient(config)

	tflog.Debug(ctx, "JumpCloud client configured")
	return client, nil
}

// JumpCloudClient is an interface for interaction with the JumpCloud API
type JumpCloudClient interface {
	DoRequest(method, path string, body interface{}) ([]byte, error)
	GetApiKey() string
	GetOrgID() string
}
