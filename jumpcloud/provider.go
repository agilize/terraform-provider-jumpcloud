package jumpcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	// For now, keep only the necessary imports and comment out the rest to avoid linter errors
	// Uncomment as needed when implementing new resources
	//"registry.terraform.io/agilize/jumpcloud/jumpcloud/admin"
	//"registry.terraform.io/agilize/jumpcloud/jumpcloud/appcatalog"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/authentication"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/iplist"

	//"registry.terraform.io/agilize/jumpcloud/jumpcloud/password_policies"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/radius"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/scim"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/system_groups"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/user_associations"

	//users "registry.terraform.io/agilize/jumpcloud/jumpcloud/user_groups"
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
			//"jumpcloud_appcatalog_application": appcatalog.ResourceAppCatalogApplication(),
			//"jumpcloud_appcatalog_assignment":  appcatalog.ResourceAssignment(),
			//"jumpcloud_appcatalog_category":    appcatalog.ResourceCategory(),

			// Legacy resource names - will be deprecated in future versions
			//"jumpcloud_app_catalog_application": appcatalog.ResourceAppCatalogApplication(),
			//"jumpcloud_app_catalog_assignment":  appcatalog.ResourceAssignment(),
			//"jumpcloud_app_catalog_category":    appcatalog.ResourceCategory(),

			// Authentication - Resources
			"jumpcloud_auth_policy":             authentication.ResourcePolicy(),
			"jumpcloud_auth_policy_binding":     authentication.ResourcePolicyBinding(),
			"jumpcloud_conditional_access_rule": authentication.ResourceConditionalAccessRule(),

			// IP Lists - Resources
			"jumpcloud_ip_list":            iplist.ResourceList(),
			"jumpcloud_ip_list_assignment": iplist.ResourceListAssignment(),

			// Platform Administrators - Resources
			//"jumpcloud_admin_user":         admin.ResourceUser(),
			//"jumpcloud_admin_role":         admin.ResourceRole(),
			//"jumpcloud_admin_role_binding": admin.ResourceRoleBinding(),

			// Password Policies - Resources
			//"jumpcloud_password_policy": password_policies.ResourcePasswordPolicy(),

			// User Group Resources
			//"jumpcloud_user_group_membership": users.ResourceMembership(),

			// User Association Resources
			"jumpcloud_user_system_association": user_associations.ResourceSystem(),

			// System Group Resources
			"jumpcloud_system_group":            system_groups.ResourceGroup(),
			"jumpcloud_system_group_membership": system_groups.ResourceMembership(),

			// RADIUS Resources
			"jumpcloud_radius_server": radius.ResourceServer(),

			// SCIM Resources
			"jumpcloud_scim_server":            scim.ResourceServer(),
			"jumpcloud_scim_attribute_mapping": scim.ResourceAttributeMapping(),
			"jumpcloud_scim_integration":       scim.ResourceIntegration(),

			// TODO: Move the remaining resources to their appropriate domain packages
			// and update the imports here
		},
		DataSourcesMap: map[string]*schema.Resource{
			// App Catalog - Data Sources
			//"jumpcloud_appcatalog_applications": appcatalog.DataSourceAppCatalogApplications(),
			//"jumpcloud_appcatalog_categories":   appcatalog.DataSourceCategories(),

			// Legacy data source names - will be deprecated in future versions
			//"jumpcloud_app_catalog_applications": appcatalog.DataSourceAppCatalogApplications(),
			//"jumpcloud_app_catalog_categories":   appcatalog.DataSourceCategories(),

			// Authentication - Data Sources
			"jumpcloud_auth_policy_templates": authentication.DataSourcePolicyTemplates(),
			"jumpcloud_auth_policies":         authentication.DataSourcePolicies(),

			// IP Lists - Data Sources
			"jumpcloud_ip_lists":     iplist.DataSourceLists(),
			"jumpcloud_ip_locations": iplist.DataSourceLocations(),

			// Platform Administrators - Data Sources
			//"jumpcloud_admin_users": admin.DataSourceUsers(),

			// SCIM - Data Sources
			"jumpcloud_scim_servers": scim.DataSourceServers(),
			"jumpcloud_scim_schema":  scim.DataSourceSchema(),

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
