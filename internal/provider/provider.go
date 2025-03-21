package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/admin"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/alerts"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/appcatalog"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/authentication"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/iplist"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/metrics"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/systems"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/users"
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
			"jumpcloud_user":       users.ResourceUser(),
			"jumpcloud_user_group": users.ResourceUserGroup(),

			// System resources
			"jumpcloud_system": systems.ResourceSystem(),

			// MDM - Resources
			"jumpcloud_mdm_configuration":      resourceMDMConfiguration(),
			"jumpcloud_mdm_policy":             resourceMDMPolicy(),
			"jumpcloud_mdm_profile":            resourceMDMProfile(),
			"jumpcloud_mdm_enrollment_profile": resourceMDMEnrollmentProfile(),

			// App Catalog - Resources
			"jumpcloud_app_catalog_application": appcatalog.ResourceAppCatalogApplication(),
			"jumpcloud_app_catalog_category":    appcatalog.ResourceCategory(),
			"jumpcloud_app_catalog_assignment":  appcatalog.ResourceAssignment(),

			// Platform Administrators - Resources
			"jumpcloud_admin_user":         admin.ResourceUser(),
			"jumpcloud_admin_role":         admin.ResourceRole(),
			"jumpcloud_admin_role_binding": admin.ResourceRoleBinding(),

			// Authentication Policies - Resources
			"jumpcloud_auth_policy":             authentication.ResourcePolicy(),
			"jumpcloud_auth_policy_binding":     authentication.ResourcePolicyBinding(),
			"jumpcloud_conditional_access_rule": authentication.ResourceConditionalAccessRule(),

			// IP Lists - Resources
			"jumpcloud_ip_list":            iplist.ResourceList(),
			"jumpcloud_ip_list_assignment": iplist.ResourceListAssignment(),

			// Alerts & Notifications - Resources
			"jumpcloud_alert_configuration":    alerts.ResourceAlertConfiguration(),
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
			"jumpcloud_auth_policies":         authentication.DataSourcePolicies(),
			"jumpcloud_auth_policy_templates": authentication.DataSourcePolicyTemplates(),

			// IP Lists - Data Sources
			"jumpcloud_ip_lists":     iplist.DataSourceLists(),
			"jumpcloud_ip_locations": iplist.DataSourceLocations(),

			// Alerts & Notifications - Data Sources
			"jumpcloud_alerts":          alerts.DataSourceAlerts(),
			"jumpcloud_alert_templates": alerts.DataSourceAlertTemplates(),

			// Metrics - Data Sources
			"jumpcloud_system_metrics": metrics.DataSourceSystemMetrics(),

			// Other Data Sources
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
