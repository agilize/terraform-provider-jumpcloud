package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/internal/provider/resources/mdm"
	"registry.terraform.io/agilize/jumpcloud/internal/provider/resources/system"
	"registry.terraform.io/agilize/jumpcloud/internal/provider/resources/user"
	"registry.terraform.io/agilize/jumpcloud/internal/provider/resources/usergroup"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/admin"
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
			// User resources
			"jumpcloud_user": user.ResourceUser(),

			// System resources
			"jumpcloud_system": system.ResourceSystem(),

			// User Group resources
			"jumpcloud_user_group": usergroup.ResourceUserGroup(),

			// MDM - Resources
			"jumpcloud_mdm_configuration":      mdm.ResourceMDMConfiguration(),
			"jumpcloud_mdm_policy":             mdm.ResourceMDMPolicy(),
			"jumpcloud_mdm_profile":            mdm.ResourceMDMProfile(),
			"jumpcloud_mdm_enrollment_profile": mdm.ResourceMDMEnrollmentProfile(),

			// App Catalog - Resources
			"jumpcloud_app_catalog_application": appcatalog.ResourceAppCatalogApplication(),
			"jumpcloud_app_catalog_category":    appcatalog.ResourceCategory(),
			"jumpcloud_app_catalog_assignment":  appcatalog.ResourceAssignment(),

			// Platform Administrators - Resources
			"jumpcloud_admin_user":         admin.ResourceUser(),
			"jumpcloud_admin_role":         admin.ResourceRole(),
			"jumpcloud_admin_role_binding": admin.ResourceRoleBinding(),

			// Authentication Policies - Resources
			"jumpcloud_auth_policy":             resourceAuthPolicy(),
			"jumpcloud_auth_policy_binding":     resourceAuthPolicyBinding(),
			"jumpcloud_conditional_access_rule": resourceConditionalAccessRule(),

			// IP Lists - Resources
			"jumpcloud_ip_list":            resourceIPList(),
			"jumpcloud_ip_list_assignment": resourceIPListAssignment(),

			// New resources
			"jumpcloud_alert_configuration":    resourceAlertConfiguration(),
			"jumpcloud_notification_channel":   resourceNotificationChannel(),
			"jumpcloud_monitoring_threshold":   resourceMonitoringThreshold(),
			"jumpcloud_password_policy":        resourcePasswordPolicy(),
			"jumpcloud_software_update_policy": resourceSoftwareUpdatePolicy(),
			"jumpcloud_scim_server":            resourceScimServer(),
			"jumpcloud_active_directory":       resourceActiveDirectory(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			// MDM - Data Sources
			"jumpcloud_mdm_devices":  dataSourceMDMDevices(),
			"jumpcloud_mdm_policies": dataSourceMDMPolicies(),
			"jumpcloud_mdm_stats":    dataSourceMDMStats(),

			// App Catalog - Data Sources
			"jumpcloud_app_catalog_applications": appcatalog.DataSourceAppCatalogApplications(),
			"jumpcloud_app_catalog_application":  appcatalog.DataSourceApplication(),

			// Platform Administrators - Data Sources
			"jumpcloud_admin_users":      admin.DataSourceUsers(),
			"jumpcloud_admin_roles":      admin.DataSourceRoles(),
			"jumpcloud_admin_audit_logs": admin.DataSourceAuditLogs(),

			// Authentication Policies - Data Sources
			"jumpcloud_auth_policies":         dataSourceAuthPolicies(),
			"jumpcloud_auth_policy_templates": dataSourceAuthPolicyTemplates(),

			// IP Lists - Data Sources
			"jumpcloud_ip_lists":     dataSourceIPLists(),
			"jumpcloud_ip_locations": dataSourceIPLocations(),

			// New data sources
			"jumpcloud_alerts":                   dataSourceAlerts(),
			"jumpcloud_alert_templates":          dataSourceAlertTemplates(),
			"jumpcloud_system_metrics":           dataSourceSystemMetrics(),
			"jumpcloud_scim_servers":             dataSourceScimServers(),
			"jumpcloud_software_packages":        dataSourceSoftwarePackages(),
			"jumpcloud_software_update_policies": dataSourceSoftwareUpdatePolicies(),
			"jumpcloud_active_directories":       dataSourceActiveDirectories(),
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
