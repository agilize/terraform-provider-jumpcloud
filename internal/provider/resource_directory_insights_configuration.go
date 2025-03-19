package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// DirectoryInsightsConfig representa a configuração do Directory Insights no JumpCloud
type DirectoryInsightsConfig struct {
	ID                    string   `json:"_id,omitempty"`
	OrgID                 string   `json:"orgId,omitempty"`
	RetentionDays         int      `json:"retentionDays"`
	EnabledEventTypes     []string `json:"enabledEventTypes"`
	ExportToCloudWatch    bool     `json:"exportToCloudWatch"`
	ExportToDatadog       bool     `json:"exportToDatadog"`
	DatadogRegion         string   `json:"datadogRegion,omitempty"`
	DatadogAPIKey         string   `json:"datadogApiKey,omitempty"`
	EnabledAlertingEvents []string `json:"enabledAlertingEvents,omitempty"`
	NotificationEmails    []string `json:"notificationEmails,omitempty"`
}

// resourceDirectoryInsightsConfiguration retorna o recurso para gerenciar a configuração do Directory Insights
func resourceDirectoryInsightsConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDirectoryInsightsConfigurationCreate,
		ReadContext:   resourceDirectoryInsightsConfigurationRead,
		UpdateContext: resourceDirectoryInsightsConfigurationUpdate,
		DeleteContext: resourceDirectoryInsightsConfigurationDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"retention_days": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 90),
				Description:  "Número de dias para retenção dos eventos (1-90)",
			},
			"enabled_event_types": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Lista de tipos de eventos habilitados para coleta",
			},
			"export_to_cloudwatch": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Se os eventos devem ser exportados para o AWS CloudWatch",
			},
			"export_to_datadog": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Se os eventos devem ser exportados para o Datadog",
			},
			"datadog_region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Região do Datadog para exportação de eventos",
			},
			"datadog_api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Chave de API do Datadog para exportação de eventos",
			},
			"enabled_alerting_events": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Lista de tipos de eventos para os quais alertas serão enviados",
			},
			"notification_emails": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Lista de emails para receber notificações de alertas",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// resourceDirectoryInsightsConfigurationCreate cria uma nova configuração do Directory Insights
func resourceDirectoryInsightsConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir configuração
	config := &DirectoryInsightsConfig{
		RetentionDays:      d.Get("retention_days").(int),
		ExportToCloudWatch: d.Get("export_to_cloudwatch").(bool),
		ExportToDatadog:    d.Get("export_to_datadog").(bool),
	}

	// Campos opcionais
	if v, ok := d.GetOk("org_id"); ok {
		config.OrgID = v.(string)
	}

	if v, ok := d.GetOk("datadog_region"); ok {
		config.DatadogRegion = v.(string)
	}

	if v, ok := d.GetOk("datadog_api_key"); ok {
		config.DatadogAPIKey = v.(string)
	}

	// Processar listas
	if v, ok := d.GetOk("enabled_event_types"); ok {
		eventTypes := v.([]interface{})
		config.EnabledEventTypes = make([]string, len(eventTypes))
		for i, eventType := range eventTypes {
			config.EnabledEventTypes[i] = eventType.(string)
		}
	}

	if v, ok := d.GetOk("enabled_alerting_events"); ok {
		alertingEvents := v.([]interface{})
		config.EnabledAlertingEvents = make([]string, len(alertingEvents))
		for i, event := range alertingEvents {
			config.EnabledAlertingEvents[i] = event.(string)
		}
	}

	if v, ok := d.GetOk("notification_emails"); ok {
		emails := v.([]interface{})
		config.NotificationEmails = make([]string, len(emails))
		for i, email := range emails {
			config.NotificationEmails[i] = email.(string)
		}
	}

	// Serializar para JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configuração do Directory Insights: %v", err))
	}

	// Criar configuração via API
	tflog.Debug(ctx, "Criando configuração do Directory Insights")
	resp, err := c.DoRequest(http.MethodPost, "/insights/directory/v1/config", configJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar configuração do Directory Insights: %v", err))
	}

	// Deserializar resposta
	var createdConfig DirectoryInsightsConfig
	if err := json.Unmarshal(resp, &createdConfig); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdConfig.ID == "" {
		return diag.FromErr(fmt.Errorf("configuração do Directory Insights criada sem ID"))
	}

	d.SetId(createdConfig.ID)
	return resourceDirectoryInsightsConfigurationRead(ctx, d, meta)
}

// resourceDirectoryInsightsConfigurationRead lê os detalhes da configuração do Directory Insights
func resourceDirectoryInsightsConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da configuração do Directory Insights não fornecido"))
	}

	// Buscar configuração via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo configuração do Directory Insights com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, "/insights/directory/v1/config", nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Configuração do Directory Insights %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler configuração do Directory Insights: %v", err))
	}

	// Deserializar resposta
	var config DirectoryInsightsConfig
	if err := json.Unmarshal(resp, &config); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("retention_days", config.RetentionDays)
	d.Set("export_to_cloudwatch", config.ExportToCloudWatch)
	d.Set("export_to_datadog", config.ExportToDatadog)
	d.Set("datadog_region", config.DatadogRegion)

	// Não definimos datadog_api_key para evitar expor credenciais sensíveis

	if config.OrgID != "" {
		d.Set("org_id", config.OrgID)
	}

	if config.EnabledEventTypes != nil {
		d.Set("enabled_event_types", config.EnabledEventTypes)
	}

	if config.EnabledAlertingEvents != nil {
		d.Set("enabled_alerting_events", config.EnabledAlertingEvents)
	}

	if config.NotificationEmails != nil {
		d.Set("notification_emails", config.NotificationEmails)
	}

	return diags
}

// resourceDirectoryInsightsConfigurationUpdate atualiza uma configuração existente do Directory Insights
func resourceDirectoryInsightsConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da configuração do Directory Insights não fornecido"))
	}

	// Construir configuração atualizada
	config := &DirectoryInsightsConfig{
		ID:                 d.Id(),
		RetentionDays:      d.Get("retention_days").(int),
		ExportToCloudWatch: d.Get("export_to_cloudwatch").(bool),
		ExportToDatadog:    d.Get("export_to_datadog").(bool),
	}

	// Campos opcionais
	if v, ok := d.GetOk("org_id"); ok {
		config.OrgID = v.(string)
	}

	if v, ok := d.GetOk("datadog_region"); ok {
		config.DatadogRegion = v.(string)
	}

	if v, ok := d.GetOk("datadog_api_key"); ok {
		config.DatadogAPIKey = v.(string)
	}

	// Processar listas
	if v, ok := d.GetOk("enabled_event_types"); ok {
		eventTypes := v.([]interface{})
		config.EnabledEventTypes = make([]string, len(eventTypes))
		for i, eventType := range eventTypes {
			config.EnabledEventTypes[i] = eventType.(string)
		}
	}

	if v, ok := d.GetOk("enabled_alerting_events"); ok {
		alertingEvents := v.([]interface{})
		config.EnabledAlertingEvents = make([]string, len(alertingEvents))
		for i, event := range alertingEvents {
			config.EnabledAlertingEvents[i] = event.(string)
		}
	}

	if v, ok := d.GetOk("notification_emails"); ok {
		emails := v.([]interface{})
		config.NotificationEmails = make([]string, len(emails))
		for i, email := range emails {
			config.NotificationEmails[i] = email.(string)
		}
	}

	// Serializar para JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configuração do Directory Insights: %v", err))
	}

	// Atualizar configuração via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando configuração do Directory Insights: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/insights/directory/v1/config/%s", id), configJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar configuração do Directory Insights: %v", err))
	}

	// Deserializar resposta
	var updatedConfig DirectoryInsightsConfig
	if err := json.Unmarshal(resp, &updatedConfig); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceDirectoryInsightsConfigurationRead(ctx, d, meta)
}

// resourceDirectoryInsightsConfigurationDelete desativa a configuração do Directory Insights
func resourceDirectoryInsightsConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Nota: Em muitos casos, configurações não são realmente excluídas, mas sim desativadas ou resetadas.
	// Aqui estamos assumindo que podemos desativar o Directory Insights com uma configuração mínima.

	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da configuração do Directory Insights não fornecido"))
	}

	// Configuração mínima (desativada)
	config := &DirectoryInsightsConfig{
		ID:                 id,
		RetentionDays:      1,          // Mínimo
		EnabledEventTypes:  []string{}, // Nenhum evento habilitado
		ExportToCloudWatch: false,
		ExportToDatadog:    false,
	}

	// Serializar para JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configuração do Directory Insights: %v", err))
	}

	// Desativar configuração via API
	tflog.Debug(ctx, fmt.Sprintf("Desativando configuração do Directory Insights: %s", id))
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/insights/directory/v1/config/%s", id), configJSON)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Configuração do Directory Insights %s não encontrada, considerando excluída", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao desativar configuração do Directory Insights: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
