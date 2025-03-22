package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"registry.terraform.io/agilize/jumpcloud/jumpcloud/alerts"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/appcatalog"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/device_management/software_management"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/organization/webhooks"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/systems"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/user_authentication/sso_applications/sso"
	usergroups "registry.terraform.io/agilize/jumpcloud/jumpcloud/user_groups"
	users "registry.terraform.io/agilize/jumpcloud/jumpcloud/user_management/users"
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
			"jumpcloud_user":                  users.ResourceUser(),
			"jumpcloud_user_group":            usergroups.ResourceUserGroup(),
			"jumpcloud_user_group_membership": usergroups.ResourceMembership(),

			// System resources
			"jumpcloud_system": systems.ResourceSystem(),

			// Alert resources
			"jumpcloud_alert_configuration": alerts.ResourceAlertConfiguration(),

			// Webhook resources
			"jumpcloud_webhook":              webhooks.ResourceWebhook(),
			"jumpcloud_webhook_subscription": webhooks.ResourceWebhookSubscription(),

			// Application Catalog - Resources
			"jumpcloud_app_catalog_category":   appcatalog.ResourceCategory(),
			"jumpcloud_app_catalog_assignment": appcatalog.ResourceAssignment(),

			// Software resources
			"jumpcloud_software_package":       software_management.ResourceSoftwarePackage(),
			"jumpcloud_software_update_policy": software_management.ResourceSoftwareUpdatePolicy(),
			"jumpcloud_software_deployment":    software_management.ResourceSoftwareDeployment(),

			// SSO Application resources
			"jumpcloud_sso_application": sso.ResourceSSOApplication(),

			// Other resources that haven't been refactored yet
			"jumpcloud_user_system_association":          resourceUserSystemAssociation(),
			"jumpcloud_system_group":                     resourceSystemGroup(),
			"jumpcloud_system_group_membership":          resourceSystemGroupMembership(),
			"jumpcloud_scim_attribute_mapping":           resourceScimAttributeMapping(),
			"jumpcloud_scim_integration":                 resourceScimIntegration(),
			"jumpcloud_scim_server":                      resourceScimServer(),
			"jumpcloud_policy":                           resourcePolicy(),
			"jumpcloud_policy_association":               resourcePolicyAssociation(),
			"jumpcloud_password_policy":                  resourcePasswordPolicy(),
			"jumpcloud_password_safe":                    resourcePasswordSafe(),
			"jumpcloud_password_entry":                   resourcePasswordEntry(),
			"jumpcloud_mfa_configuration":                resourceMFAConfiguration(),
			"jumpcloud_mfa_settings":                     resourceMFASettings(),
			"jumpcloud_radius_server":                    resourceRadiusServer(),
			"jumpcloud_monitoring_threshold":             resourceMonitoringThreshold(),
			"jumpcloud_notification_channel":             resourceNotificationChannel(),
			"jumpcloud_mdm_configuration":                resourceMDMConfiguration(),
			"jumpcloud_mdm_enrollment_profile":           resourceMDMEnrollmentProfile(),
			"jumpcloud_mdm_policy":                       resourceMDMPolicy(),
			"jumpcloud_mdm_profile":                      resourceMDMProfile(),
			"jumpcloud_command":                          resourceCommand(),
			"jumpcloud_command_association":              resourceCommandAssociation(),
			"jumpcloud_command_schedule":                 resourceCommandSchedule(),
			"jumpcloud_organization":                     resourceOrganization(),
			"jumpcloud_organization_settings":            resourceOrganizationSettings(),
			"jumpcloud_directory_insights_configuration": resourceDirectoryInsightsConfiguration(),
			"jumpcloud_oauth_authorization":              resourceOAuthAuthorization(),
			"jumpcloud_oauth_user":                       resourceOAuthUser(),
			"jumpcloud_api_key":                          resourceAPIKey(),
			"jumpcloud_api_key_binding":                  resourceAPIKeyBinding(),
			"jumpcloud_active_directory":                 resourceActiveDirectory(),
			"jumpcloud_application":                      resourceApplication(),
			"jumpcloud_application_group_mapping":        resourceApplicationGroupMapping(),
			"jumpcloud_application_user_mapping":         resourceApplicationUserMapping(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			// Webhook data sources
			"jumpcloud_webhook":         webhooks.DataSourceWebhook(),
			"jumpcloud_sso_application": sso.DataSourceSSOApplication(),
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
