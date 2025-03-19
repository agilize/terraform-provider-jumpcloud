package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// New retorna uma instância do plugin do provider
func New() *schema.Provider {
	return Provider()
}

// Provider retorna um schema.Provider para o JumpCloud.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_API_KEY", nil),
				Description: "Chave de API para operações do JumpCloud.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_ORG_ID", nil),
				Description: "ID da organização para ambientes multi-tenant do JumpCloud.",
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_API_URL", "https://console.jumpcloud.com/api"),
				Description: "URL da API do JumpCloud.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			// MDM - Recursos
			"jumpcloud_mdm_configuration":      resourceMDMConfiguration(),
			"jumpcloud_mdm_policy":             resourceMDMPolicy(),
			"jumpcloud_mdm_profile":            resourceMDMProfile(),
			"jumpcloud_mdm_enrollment_profile": resourceMDMEnrollmentProfile(),

			// App Catalog - Recursos
			"jumpcloud_app_catalog_application": resourceAppCatalogApplication(),
			"jumpcloud_app_catalog_category":    resourceAppCatalogCategory(),
			"jumpcloud_app_catalog_assignment":  resourceAppCatalogAssignment(),

			// Administradores da Plataforma - Recursos
			"jumpcloud_admin_user":         resourceAdminUser(),
			"jumpcloud_admin_role":         resourceAdminRole(),
			"jumpcloud_admin_role_binding": resourceAdminRoleBinding(),

			// Políticas de Autenticação - Recursos
			"jumpcloud_auth_policy":             resourceAuthPolicy(),
			"jumpcloud_auth_policy_binding":     resourceAuthPolicyBinding(),
			"jumpcloud_conditional_access_rule": resourceConditionalAccessRule(),

			// Listas de IPs - Recursos
			"jumpcloud_ip_list":            resourceIPList(),
			"jumpcloud_ip_list_assignment": resourceIPListAssignment(),

			// Novos recursos
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
			"jumpcloud_app_catalog_applications": dataSourceAppCatalogApplications(),
			"jumpcloud_app_catalog_categories":   dataSourceAppCatalogCategories(),

			// Administradores da Plataforma - Data Sources
			"jumpcloud_admin_users":      dataSourceAdminUsers(),
			"jumpcloud_admin_roles":      dataSourceAdminRoles(),
			"jumpcloud_admin_audit_logs": dataSourceAdminAuditLogs(),

			// Políticas de Autenticação - Data Sources
			"jumpcloud_auth_policies":         dataSourceAuthPolicies(),
			"jumpcloud_auth_policy_templates": dataSourceAuthPolicyTemplates(),

			// Listas de IPs - Data Sources
			"jumpcloud_ip_lists":     dataSourceIPLists(),
			"jumpcloud_ip_locations": dataSourceIPLocations(),

			// Novos data sources
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

// providerConfigure configura o provider com detalhes de autenticação
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	tflog.Info(ctx, "Configurando cliente JumpCloud")

	apiKey := d.Get("api_key").(string)
	orgID := d.Get("org_id").(string)
	apiURL := d.Get("api_url").(string)

	c := &jumpCloudClient{
		apiKey: apiKey,
		orgID:  orgID,
		apiURL: apiURL,
	}

	tflog.Debug(ctx, "Cliente JumpCloud configurado")
	return c, nil
}

// JumpCloudClient é uma interface para interação com a API do JumpCloud
type JumpCloudClient interface {
	DoRequest(method, path string, body interface{}) ([]byte, error)
	GetApiKey() string
	GetOrgID() string
}

// jumpCloudClient implementa a interface JumpCloudClient
type jumpCloudClient struct {
	apiKey string
	orgID  string
	apiURL string
}

// DoRequest realiza uma requisição para a API do JumpCloud
func (c *jumpCloudClient) DoRequest(method, path string, body []byte) ([]byte, error) {
	// Implementação simplificada para o exemplo
	return nil, nil
}

func (c *jumpCloudClient) GetApiKey() string {
	return c.apiKey
}

func (c *jumpCloudClient) GetOrgID() string {
	return c.orgID
}
