package jumpcloud

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/pkg/apiclient"

	// Device Management
	devices_commands "registry.terraform.io/agilize/jumpcloud/jumpcloud/device_management/commands"
	device_groups "registry.terraform.io/agilize/jumpcloud/jumpcloud/device_management/device_groups"
	devices "registry.terraform.io/agilize/jumpcloud/jumpcloud/device_management/devices"
	devices_mdm "registry.terraform.io/agilize/jumpcloud/jumpcloud/device_management/mdm"
	devices_software_management "registry.terraform.io/agilize/jumpcloud/jumpcloud/device_management/software_management"

	// Insights
	insights_alerts "registry.terraform.io/agilize/jumpcloud/jumpcloud/insights/alerts"
	insights_directory_insights "registry.terraform.io/agilize/jumpcloud/jumpcloud/insights/directory_insights"

	// Organization Settings
	admin_roles "registry.terraform.io/agilize/jumpcloud/jumpcloud/organization_settings/admin_roles"
	admin_users "registry.terraform.io/agilize/jumpcloud/jumpcloud/organization_settings/admin_users"
	organization_api_keys "registry.terraform.io/agilize/jumpcloud/jumpcloud/organization_settings/api_keys"
	organization_audit_logs "registry.terraform.io/agilize/jumpcloud/jumpcloud/organization_settings/audit_logs"
	organization_metrics "registry.terraform.io/agilize/jumpcloud/jumpcloud/organization_settings/metrics"
	organization_monitors "registry.terraform.io/agilize/jumpcloud/jumpcloud/organization_settings/monitors"
	organization_notifications "registry.terraform.io/agilize/jumpcloud/jumpcloud/organization_settings/notifications"
	organization_settings "registry.terraform.io/agilize/jumpcloud/jumpcloud/organization_settings/settings"
	organization_webhooks "registry.terraform.io/agilize/jumpcloud/jumpcloud/organization_settings/webhooks"

	// Security Management
	authentication "registry.terraform.io/agilize/jumpcloud/jumpcloud/security_management/attempts"
	security_conditional_access "registry.terraform.io/agilize/jumpcloud/jumpcloud/security_management/conditional_access"
	security_iplist "registry.terraform.io/agilize/jumpcloud/jumpcloud/security_management/iplist"
	security_mfa "registry.terraform.io/agilize/jumpcloud/jumpcloud/security_management/mfa"
	security_password_policies "registry.terraform.io/agilize/jumpcloud/jumpcloud/security_management/password_policies"

	// User Authentication
	password_manager "registry.terraform.io/agilize/jumpcloud/jumpcloud/user_authentication/password_manager"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/user_authentication/radius"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/user_authentication/scim"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/user_authentication/sso"

	// User Management
	usergroups "registry.terraform.io/agilize/jumpcloud/jumpcloud/user_management/user_groups"
	users "registry.terraform.io/agilize/jumpcloud/jumpcloud/user_management/users"
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
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_API_URL", "https://console.jumpcloud.com"),
				Description: "JumpCloud API URL.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			// Organization Settings - Admin Users
			"jumpcloud_admin_user": admin_users.ResourceUser(),

			// Organization Settings - Admin Roles
			"jumpcloud_admin_role":         admin_roles.ResourceRole(),
			"jumpcloud_admin_role_binding": admin_roles.ResourceRoleBinding(),

			// User Authentication - SCIM
			"jumpcloud_application_scim_server":            scim.ResourceServer(),
			"jumpcloud_application_scim_attribute_mapping": scim.ResourceAttributeMapping(),
			"jumpcloud_application_scim_integration":       scim.ResourceIntegration(),

			// User Authentication - SSO
			"jumpcloud_application_sso_application": sso.ResourceSSOApplication(),

			// Security Management - Conditional Access
			"jumpcloud_authentication_conditional_access_rule": security_conditional_access.ResourceConditionalAccessRule(),

			// Security Management - IP Lists
			"jumpcloud_authentication_ip_list":            security_iplist.ResourceList(),
			"jumpcloud_authentication_ip_list_assignment": security_iplist.ResourceListAssignment(),

			// Security Management - MFA
			"jumpcloud_authentication_mfa_configuration": security_mfa.ResourceConfiguration(),
			"jumpcloud_authentication_mfa_settings":      security_mfa.ResourceSettings(),

			// Security Management - Password Policies
			"jumpcloud_password_policy": security_password_policies.ResourcePasswordPolicy(),

			// User Authentication - RADIUS
			"jumpcloud_authentication_radius_server": radius.ResourceServer(),

			// Devices Commands - Resources
			"jumpcloud_devices_command":             devices_commands.ResourceCommand(),
			"jumpcloud_devices_command_association": devices_commands.ResourceCommandAssociation(),
			"jumpcloud_devices_command_schedule":    devices_commands.ResourceCommandSchedule(),

			// Device Groups - Resources
			"jumpcloud_devices_group":            device_groups.ResourceDeviceGroup(),
			"jumpcloud_devices_group_membership": device_groups.ResourceDeviceGroupMembership(),

			// Devices - Resources
			"jumpcloud_devices": devices.ResourceDevice(),

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

			// Organization Settings - API Keys
			"jumpcloud_organization_api_key":         organization_api_keys.ResourceKey(),
			"jumpcloud_organization_api_key_binding": organization_api_keys.ResourceKeyBinding(),

			// Organization Settings - Monitoring
			"jumpcloud_organization_monitoring_threshold": organization_monitors.ResourceThreshold(),

			// Organization Settings - Notifications
			"jumpcloud_organization_notification_channel": organization_notifications.ResourceChannel(),

			// Organization Settings - Settings
			"jumpcloud_organization":          organization_settings.ResourceOrganization(),
			"jumpcloud_organization_settings": organization_settings.ResourceSettings(),

			// Organization Settings - Webhooks
			"jumpcloud_organization_webhook":              organization_webhooks.ResourceWebhook(),
			"jumpcloud_organization_webhook_subscription": organization_webhooks.ResourceWebhookSubscription(),

			// Insights - Directory Insights
			"jumpcloud_directory_insights_configuration": insights_directory_insights.ResourceConfiguration(),

			// User Authentication - Password Manager
			"jumpcloud_password_safe":  password_manager.ResourceSafe(),
			"jumpcloud_password_entry": password_manager.ResourceEntry(),

			// User Management - User Groups
			"jumpcloud_user_group":                         usergroups.ResourceUserGroup(),
			"jumpcloud_user_group_membership":              usergroups.ResourceUserGroupMembership(),
			"jumpcloud_user_group_application_association": usergroups.ResourceUserGroupApplicationAssociation(),

			// User Management - Users
			"jumpcloud_user":                         users.ResourceUser(),
			"jumpcloud_user_device_association":      users.ResourceUserDeviceAssociation(),
			"jumpcloud_user_application_association": users.ResourceUserApplicationAssociation(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			// Organization Settings - Admin Roles
			"jumpcloud_admin_roles": admin_roles.DataSourceRoles(),

			// Organization Settings - Admin Users
			"jumpcloud_admin_users": admin_users.DataSourceUsers(),

			// User Management - User Groups
			"jumpcloud_user_group":                         usergroups.DataSourceUserGroup(),
			"jumpcloud_user_group_membership":              usergroups.DataSourceUserGroupMembership(),
			"jumpcloud_user_group_application_association": usergroups.DataSourceUserGroupApplicationAssociation(),

			// User Management - Users
			"jumpcloud_user":                         users.DataSourceUser(),
			"jumpcloud_user_device_association":      users.DataSourceUserDeviceAssociation(),
			"jumpcloud_user_application_association": users.DataSourceUserApplicationAssociation(),

			// User Authentication - SCIM
			"jumpcloud_application_scim_servers": scim.DataSourceServers(),
			"jumpcloud_application_scim_schema":  scim.DataSourceSchema(),

			// User Authentication - SSO
			"jumpcloud_application_sso_application": sso.DataSourceSSOApplication(),

			// Security Management - Attempts
			"jumpcloud_authentication_attempts": authentication.DataSourceAttempts(),

			// Security Management - IP Lists
			"jumpcloud_authentication_ip_lists":     security_iplist.DataSourceLists(),
			"jumpcloud_authentication_ip_locations": security_iplist.DataSourceLocations(),

			// Security Management - MFA
			"jumpcloud_authentication_mfa_settings": security_mfa.DataSourceSettings(),
			"jumpcloud_authentication_mfa_stats":    security_mfa.DataSourceStats(),

			// User Authentication - RADIUS
			"jumpcloud_authentication_radius_server": radius.DataSourceServer(),

			// Devices Commands - Data Sources
			"jumpcloud_devices_command": devices_commands.DataSourceCommand(),

			// Devices System - Data Sources
			"jumpcloud_devices_group": device_groups.DataSourceDeviceGroup(),
			"jumpcloud_devices":       devices.DataSourceDevice(),

			// Devices MDM - Data Sources
			"jumpcloud_devices_mdm_stats":    devices_mdm.DataSourceStats(),
			"jumpcloud_devices_mdm_devices":  devices_mdm.DataSourceDevices(),
			"jumpcloud_devices_mdm_policies": devices_mdm.DataSourcePolicies(),

			// Devices Software Management - Data Sources
			"jumpcloud_devices_software_packages":          devices_software_management.DataSourceSoftwarePackages(),
			"jumpcloud_devices_software_update_policies":   devices_software_management.DataSourceSoftwareUpdatePolicies(),
			"jumpcloud_devices_software_deployment_status": devices_software_management.DataSourceSoftwareDeploymentStatus(),

			// Insights - Directory Insights
			"jumpcloud_directory_insights_events": insights_directory_insights.DataSourceEvents(),

			// Insights - Alerts
			"jumpcloud_organization_alerts":          insights_alerts.DataSourceAlerts(),
			"jumpcloud_organization_alert_templates": insights_alerts.DataSourceAlertTemplates(),

			// Organization Settings - Audit Logs
			"jumpcloud_organization_audit_logs": organization_audit_logs.DataSourceAuditLogs(),

			// Organization Settings - Metrics
			"jumpcloud_organization_system_metrics": organization_metrics.DataSourceSystemMetrics(),

			// Organization Settings - Webhooks
			"jumpcloud_organization_webhook": organization_webhooks.DataSourceWebhook(),

			// User Authentication - Password Manager
			"jumpcloud_password_safes": password_manager.DataSourceSafes(),

			// Security Management - Password Policies
			"jumpcloud_password_policies": security_password_policies.DataSourcePolicies(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

// providerConfigure configures the provider with authentication details
func providerConfigure(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	tflog.Info(ctx, "Configuring JumpCloud client")

	apiKey := d.Get("api_key").(string)
	orgID := d.Get("org_id").(string)
	apiURL := d.Get("api_url").(string)

	config := &apiclient.Config{
		APIKey: apiKey,
		OrgID:  orgID,
		APIURL: apiURL,
	}

	apiClient := apiclient.NewClient(config)

	// Wrap the API client with an adapter that implements the ClientInterface
	client := &clientAdapter{apiClient: apiClient}

	tflog.Debug(ctx, "JumpCloud client configured")
	return client, nil
}

// clientAdapter adapts the apiclient.Client to implement multiple client interfaces
// It implements both common.ClientInterface (with []byte) and common.APIClientInterface (with interface{})
type clientAdapter struct {
	apiClient *apiclient.Client
}

// doRequestInternal is the internal implementation that handles all request types
func (a *clientAdapter) doRequestInternal(method, path string, body interface{}) ([]byte, error) {
	var requestBody any

	// Handle different body types
	switch v := body.(type) {
	case []byte:
		// If body is already []byte, try to unmarshal it
		if len(v) > 0 {
			if err := json.Unmarshal(v, &requestBody); err != nil {
				// If unmarshal fails, use raw bytes
				bodyStr := string(v)
				if len(bodyStr) > 0 && (bodyStr[0] == '{' || bodyStr[0] == '[') {
					tflog.Warn(context.Background(), fmt.Sprintf("Failed to unmarshal JSON request body: %v. Using raw bytes.", err))
				}
				requestBody = v
			}
		}
	case nil:
		// No body
		requestBody = nil
	default:
		// For any other type, use it directly
		requestBody = v
	}

	// Log the request for debugging
	fmt.Printf("DEBUG: Making API request: %s %s\n", method, path)
	if requestBody != nil {
		bodyBytes, _ := json.Marshal(requestBody)
		fmt.Printf("DEBUG: Request body: %s\n", string(bodyBytes))
	}

	// Call the underlying API client
	result, err := a.apiClient.DoRequest(method, path, requestBody)

	// Log the response for debugging
	if err != nil {
		fmt.Printf("DEBUG: API request failed: %v\n", err)
	} else {
		fmt.Printf("DEBUG: API response: %s\n", string(result))
	}

	return result, err
}

// DoRequest implements BOTH interfaces by accepting interface{} which is compatible with []byte
// This works because []byte satisfies interface{}, and we handle the conversion internally
func (a *clientAdapter) DoRequest(method, path string, body interface{}) ([]byte, error) {
	return a.doRequestInternal(method, path, body)
}

// DoRequestWithContext implements the ClientInterface method with context support
func (a *clientAdapter) DoRequestWithContext(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	// For now, we ignore the context and delegate to DoRequest
	// TODO: Add proper context support to the underlying API client
	return a.doRequestInternal(method, path, body)
}

// GetApiKey implements the ClientInterface method with the correct signature
func (a *clientAdapter) GetApiKey() string {
	return a.apiClient.GetApiKey()
}

// GetAPIKey implements the APIClientInterface method (note the capitalization)
func (a *clientAdapter) GetAPIKey() string {
	return a.apiClient.GetApiKey()
}

// GetOrgID implements the ClientInterface method with the correct signature
func (a *clientAdapter) GetOrgID() string {
	return a.apiClient.GetOrgID()
}

// JumpCloudClient is an interface for interaction with the JumpCloud API
type JumpCloudClient interface {
	DoRequest(method, path string, body any) ([]byte, error)
	GetApiKey() string
	GetOrgID() string
}
