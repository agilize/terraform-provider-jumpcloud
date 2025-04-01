package jumpcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/pkg/apiclient"

	// Admin - Resources
	admin_roles "registry.terraform.io/agilize/jumpcloud/jumpcloud/admin/admin_roles"
	admin_users "registry.terraform.io/agilize/jumpcloud/jumpcloud/admin/admin_users"
	// Application - Resources
	application_catalog "registry.terraform.io/agilize/jumpcloud/jumpcloud/application/catalog"
	application_mappings "registry.terraform.io/agilize/jumpcloud/jumpcloud/application/mappings"
	application_oauth "registry.terraform.io/agilize/jumpcloud/jumpcloud/application/oauth"
	application_scim "registry.terraform.io/agilize/jumpcloud/jumpcloud/application/scim"
	application_sso "registry.terraform.io/agilize/jumpcloud/jumpcloud/application/sso"
	// Authentication - Resources
	authentication_attempts "registry.terraform.io/agilize/jumpcloud/jumpcloud/authentication/attempts"
	authentication_conditional_access "registry.terraform.io/agilize/jumpcloud/jumpcloud/authentication/conditional_access"
	authentication_iplist "registry.terraform.io/agilize/jumpcloud/jumpcloud/authentication/iplist"
	authentication_mfa "registry.terraform.io/agilize/jumpcloud/jumpcloud/authentication/mfa"
	authentication_policies "registry.terraform.io/agilize/jumpcloud/jumpcloud/authentication/policies"
	authentication_radius "registry.terraform.io/agilize/jumpcloud/jumpcloud/authentication/radius"
	// Devices - Resources
	devices_commands "registry.terraform.io/agilize/jumpcloud/jumpcloud/devices/commands"
	devices_mdm "registry.terraform.io/agilize/jumpcloud/jumpcloud/devices/mdm"
	devices_software_management "registry.terraform.io/agilize/jumpcloud/jumpcloud/devices/software_management"
	devices "registry.terraform.io/agilize/jumpcloud/jumpcloud/devices/system_devices"
	device_groups "registry.terraform.io/agilize/jumpcloud/jumpcloud/devices/system_groups"
	// Insights - Resources
	insights_directory_insights "registry.terraform.io/agilize/jumpcloud/jumpcloud/insights/directory_insights"
	// Organization - Resources
	organization_alerts "registry.terraform.io/agilize/jumpcloud/jumpcloud/organization/alerts"
	organization_api_keys "registry.terraform.io/agilize/jumpcloud/jumpcloud/organization/api_keys"
	organization_audit_logs "registry.terraform.io/agilize/jumpcloud/jumpcloud/organization/audit_logs"
	organization_metrics "registry.terraform.io/agilize/jumpcloud/jumpcloud/organization/metrics"
	organization_monitors "registry.terraform.io/agilize/jumpcloud/jumpcloud/organization/monitors"
	organization_notifications "registry.terraform.io/agilize/jumpcloud/jumpcloud/organization/notifications"
	organization_settings "registry.terraform.io/agilize/jumpcloud/jumpcloud/organization/settings"
	organization_webhooks "registry.terraform.io/agilize/jumpcloud/jumpcloud/organization/webhooks"
	// Password - Resources
	password_manager "registry.terraform.io/agilize/jumpcloud/jumpcloud/password/password_manager"
	password_policies "registry.terraform.io/agilize/jumpcloud/jumpcloud/password/password_policies"
	// Users - Resources
	user_associations "registry.terraform.io/agilize/jumpcloud/jumpcloud/users/user_associations"
	user_groups "registry.terraform.io/agilize/jumpcloud/jumpcloud/users/user_groups"
	users_directory "registry.terraform.io/agilize/jumpcloud/jumpcloud/users/users_directory"
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
			// Admin Users - Resources
			"jumpcloud_admin_user": admin_users.ResourceUser(),

			// Admin Roles - Resources
			"jumpcloud_admin_role":         admin_roles.ResourceRole(),
			"jumpcloud_admin_role_binding": admin_roles.ResourceRoleBinding(),

			// Application Catalog - Resources
			"jumpcloud_application_catalog_application": application_catalog.ResourceAppCatalogApplication(),
			"jumpcloud_application_catalog_assignment":  application_catalog.ResourceAssignment(),
			"jumpcloud_application_catalog_category":    application_catalog.ResourceCategory(),

			// Application Mappings - Resources
			"jumpcloud_application_mapping_user":  application_mappings.ResourceUserMapping(),
			"jumpcloud_application_mapping_group": application_mappings.ResourceGroupMapping(),

			// Application OAuth resources
			"jumpcloud_application_oauth_authorization": application_oauth.ResourceAuthorization(),
			"jumpcloud_application_oauth_user":          application_oauth.ResourceUser(),

			// Application SCIM Resources
			"jumpcloud_application_scim_server":            application_scim.ResourceServer(),
			"jumpcloud_application_scim_attribute_mapping": application_scim.ResourceAttributeMapping(),
			"jumpcloud_application_scim_integration":       application_scim.ResourceIntegration(),

			// Application SSO Resources
			"jumpcloud_application_sso_application": application_sso.ResourceSSOApplication(),

			// Authentication Conditional Access - Resources
			"jumpcloud_authentication_conditional_access_rule": authentication_conditional_access.ResourceConditionalAccessRule(),

			// Authentication IP Lists - Resources
			"jumpcloud_authentication_ip_list":            authentication_iplist.ResourceList(),
			"jumpcloud_authentication_ip_list_assignment": authentication_iplist.ResourceListAssignment(),

			// Authentication MFA - Resources
			"jumpcloud_authentication_mfa_configuration": authentication_mfa.ResourceConfiguration(),
			"jumpcloud_authentication_mfa_settings":      authentication_mfa.ResourceSettings(),

			// Authentication Policies - Resources
			"jumpcloud_authentication_policy":         authentication_policies.ResourcePolicy(),
			"jumpcloud_authentication_policy_binding": authentication_policies.ResourcePolicyBinding(),

			// Authentication RADIUS - Resources
			"jumpcloud_authentication_radius_server": authentication_radius.ResourceServer(),

			// Devices Commands - Resources
			"jumpcloud_devices_command":             devices_commands.ResourceCommand(),
			"jumpcloud_devices_command_association": devices_commands.ResourceCommandAssociation(),
			"jumpcloud_devices_command_schedule":    devices_commands.ResourceCommandSchedule(),

			// Device Groups - Resources
			"jumpcloud_devices_group":            device_groups.ResourceGroup(),
			"jumpcloud_devices_group_membership": device_groups.ResourceMembership(),

			// Devices - Resources
			"jumpcloud_devices": devices.ResourceSystem(),

			// Devices MDM - Resources
			"jumpcloud_devices_mdm_configuration":      devices_mdm.ResourceConfiguration(),
			"jumpcloud_devices_mdm_enrollment_profile": devices_mdm.ResourceEnrollmentProfile(),
			"jumpcloud_devices_mdm_policy":             devices_mdm.ResourcePolicy(),
			"jumpcloud_devices_mdm_profile":            devices_mdm.ResourceProfile(),
			"jumpcloud_devices_mdm_device_action":      devices_mdm.ResourceDeviceAction(),

			// Devices Software Management - Resources
			"jumpcloud_devices_software_package":       devices_software_management.ResourceSoftwarePackage(),
			"jumpcloud_devices_software_update_policy": devices_software_management.ResourceSoftwareUpdatePolicy(),
			"jumpcloud_devices_software_deployment":    devices_software_management.ResourceSoftwareDeployment(),

			// Organization Alerts - Resources
			"jumpcloud_organization_alert_configuration": organization_alerts.ResourceAlertConfiguration(),

			// Organization API Keys - Resources
			"jumpcloud_organization_api_key":         organization_api_keys.ResourceKey(),
			"jumpcloud_organization_api_key_binding": organization_api_keys.ResourceKeyBinding(),

			// Organization Monitoring - Resources
			"jumpcloud_organization_monitoring_threshold": organization_monitors.ResourceThreshold(),

			// Organization Notifications - Resources
			"jumpcloud_organization_notification_channel": organization_notifications.ResourceChannel(),

			// Organization Settings - Resources
			"jumpcloud_organization":          organization_settings.ResourceOrganization(),
			"jumpcloud_organization_settings": organization_settings.ResourceSettings(),

			// Organization Webhooks - Resources
			"jumpcloud_organization_webhook":              organization_webhooks.ResourceWebhook(),
			"jumpcloud_organization_webhook_subscription": organization_webhooks.ResourceWebhookSubscription(),

			// Directory Insights - Resources
			"jumpcloud_directory_insights_configuration": insights_directory_insights.ResourceConfiguration(),

			// Password Manager resources
			"jumpcloud_password_safe":  password_manager.ResourceSafe(),
			"jumpcloud_password_entry": password_manager.ResourceEntry(),

			// Password Policies - Resources
			"jumpcloud_password_policy": password_policies.ResourcePasswordPolicy(),

			// User Association Resources
			"jumpcloud_user_device_association": user_associations.ResourceSystem(),

			// User Groups Resources
			"jumpcloud_user_group":            user_groups.ResourceUserGroup(),
			"jumpcloud_user_group_membership": user_groups.ResourceMembership(),

			// Users - Resources
			"jumpcloud_user": users_directory.ResourceUser(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			// Admin Roles - Data Sources
			"jumpcloud_admin_roles": admin_roles.DataSourceRoles(),

			// Admin Users - Data Sources
			"jumpcloud_admin_users": admin_users.DataSourceUsers(),

			// Application Catalog - Data Sources
			"jumpcloud_application_catalog_application":  application_catalog.DataSourceApplication(),
			"jumpcloud_application_catalog_applications": application_catalog.DataSourceAppCatalogApplications(),
			"jumpcloud_application_catalog_categories":   application_catalog.DataSourceCategories(),

			// Application OAuth - Data Sources
			"jumpcloud_application_oauth_users": application_oauth.DataSourceUsers(),

			// Application SCIM - Data Sources
			"jumpcloud_application_scim_servers": application_scim.DataSourceServers(),
			"jumpcloud_application_scim_schema":  application_scim.DataSourceSchema(),

			// Application SSO - Data Sources
			"jumpcloud_application_sso_application": application_sso.DataSourceSSOApplication(),

			// Authentication Attempts - Data Sources
			"jumpcloud_authentication_attempts": authentication_attempts.DataSourceAttempts(),

			// Authentication IP Lists - Data Sources
			"jumpcloud_authentication_ip_lists":     authentication_iplist.DataSourceLists(),
			"jumpcloud_authentication_ip_locations": authentication_iplist.DataSourceLocations(),

			// Authentication MFA - Data Sources
			"jumpcloud_authentication_mfa_settings": authentication_mfa.DataSourceSettings(),
			"jumpcloud_authentication_mfa_stats":    authentication_mfa.DataSourceStats(),

			// Authentication Policies - Data Sources
			"jumpcloud_authentication_policy_templates": authentication_policies.DataSourcePolicyTemplates(),
			"jumpcloud_authentication_policies":         authentication_policies.DataSourcePolicies(),

			// Authentication RADIUS - Data Sources
			"jumpcloud_authentication_radius_server": authentication_radius.DataSourceServer(),

			// Devices Commands - Data Sources
			"jumpcloud_devices_command": devices_commands.DataSourceCommand(),

			// Devices System - Data Sources
			"jumpcloud_devices_group": device_groups.DataSourceGroup(),
			"jumpcloud_devices":       devices.DataSourceSystem(),

			// Devices MDM - Data Sources
			"jumpcloud_devices_mdm_stats":    devices_mdm.DataSourceStats(),
			"jumpcloud_devices_mdm_devices":  devices_mdm.DataSourceDevices(),
			"jumpcloud_devices_mdm_policies": devices_mdm.DataSourcePolicies(),

			// Devices Software Management - Data Sources
			"jumpcloud_devices_software_packages":          devices_software_management.DataSourceSoftwarePackages(),
			"jumpcloud_devices_software_update_policies":   devices_software_management.DataSourceSoftwareUpdatePolicies(),
			"jumpcloud_devices_software_deployment_status": devices_software_management.DataSourceSoftwareDeploymentStatus(),

			// Directory Insights - Data Sources
			"jumpcloud_directory_insights_events": insights_directory_insights.DataSourceEvents(),

			// Organization Alerts - Data Sources
			"jumpcloud_organization_alerts":          organization_alerts.DataSourceAlerts(),
			"jumpcloud_organization_alert_templates": organization_alerts.DataSourceAlertTemplates(),

			// Organization Audit Logs - Data Sources
			"jumpcloud_organization_audit_logs": organization_audit_logs.DataSourceAuditLogs(),

			// Organization Metrics - Data Sources
			"jumpcloud_organization_system_metrics": organization_metrics.DataSourceSystemMetrics(),

			// Organization Webhooks - Data Sources
			"jumpcloud_organization_webhook": organization_webhooks.DataSourceWebhook(),

			// Password Manager data sources
			"jumpcloud_password_safes": password_manager.DataSourceSafes(),

			// Password Policies - Data Sources
			"jumpcloud_password_policies": password_policies.DataSourcePolicies(),
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
