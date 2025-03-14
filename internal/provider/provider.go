package provider

import (
	"context"

	"github.com/ferreirafav/terraform-provider-jumpcloud/internal/client"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a *schema.Provider for JumpCloud.
func New() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_API_KEY", nil),
				Description: "The API key for JumpCloud API operations.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_ORG_ID", nil),
				Description: "The organization ID for JumpCloud API operations.",
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_API_URL", "https://console.jumpcloud.com/api"),
				Description: "The URL of the JumpCloud API.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"jumpcloud_user":   resourceUser(),
			"jumpcloud_system": resourceSystem(),
			// Add more resources here as they are implemented
		},
		DataSourcesMap: map[string]*schema.Resource{
			"jumpcloud_user":   dataSourceUser(),
			"jumpcloud_system": dataSourceSystem(),
			// Add more data sources here as they are implemented
		},
		ConfigureContextFunc: providerConfigure,
	}
}

// providerConfigure configures the provider with the authentication details and returns
// a client that can be used by resources and data sources.
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	tflog.Info(ctx, "Configuring JumpCloud client")

	apiKey := d.Get("api_key").(string)
	orgID := d.Get("org_id").(string)
	apiURL := d.Get("api_url").(string)

	config := &client.Config{
		APIKey: apiKey,
		OrgID:  orgID,
		APIURL: apiURL,
	}

	c := client.NewClient(config)

	tflog.Debug(ctx, "JumpCloud client configured")

	return c, nil
}
