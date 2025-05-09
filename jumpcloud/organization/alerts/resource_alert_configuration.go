package alerts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// AlertConfiguration representa uma configuração de alerta no JumpCloud
type AlertConfiguration struct {
	ID              string                 `json:"_id,omitempty"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	Type            string                 `json:"type"` // system_metric, login_attempt, agent_status, etc.
	Enabled         bool                   `json:"enabled"`
	Conditions      map[string]interface{} `json:"conditions,omitempty"`
	Triggers        []string               `json:"triggers,omitempty"`
	NotificationIDs []string               `json:"notificationIds,omitempty"`
	Severity        string                 `json:"severity,omitempty"` // critical, high, medium, low, info
	OrgID           string                 `json:"orgId,omitempty"`
	Created         string                 `json:"created,omitempty"`
	Updated         string                 `json:"updated,omitempty"`
}

func ResourceAlertConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAlertConfigurationCreate,
		ReadContext:   resourceAlertConfigurationRead,
		UpdateContext: resourceAlertConfigurationUpdate,
		DeleteContext: resourceAlertConfigurationDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome da configuração de alerta",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição da configuração de alerta",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"system_metric", "login_attempt", "agent_status", "policy_violation",
					"device_status", "directory_sync", "user_change", "admin_action",
				}, false),
				Description: "Tipo de alerta (system_metric, login_attempt, agent_status, etc.)",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Se o alerta está ativado ou não",
			},
			"conditions": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Condições do alerta em formato JSON",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					jsonStr := val.(string)
					var js map[string]interface{}
					if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
						errs = append(errs, fmt.Errorf("%q: JSON inválido: %s", key, err))
					}
					return
				},
			},
			"triggers": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Lista de eventos que acionam o alerta",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"notification_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "IDs dos canais de notificação a serem utilizados",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"severity": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "medium",
				ValidateFunc: validation.StringInSlice([]string{"critical", "high", "medium", "low", "info"}, false),
				Description:  "Severidade do alerta (critical, high, medium, low, info)",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da configuração de alerta",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da configuração de alerta",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceAlertConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Processar as condições (string JSON para map)
	var conditions map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("conditions").(string)), &conditions); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar condições: %v", err))
	}

	// Construir configuração de alerta
	alertConfig := &AlertConfiguration{
		Name:       d.Get("name").(string),
		Type:       d.Get("type").(string),
		Enabled:    d.Get("enabled").(bool),
		Conditions: conditions,
		Severity:   d.Get("severity").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		alertConfig.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		alertConfig.OrgID = v.(string)
	}

	// Processar triggers
	if v, ok := d.GetOk("triggers"); ok {
		triggerSet := v.(*schema.Set).List()
		triggers := make([]string, len(triggerSet))
		for i, t := range triggerSet {
			triggers[i] = t.(string)
		}
		alertConfig.Triggers = triggers
	}

	// Processar notification_ids
	if v, ok := d.GetOk("notification_ids"); ok {
		notificationSet := v.(*schema.Set).List()
		notificationIDs := make([]string, len(notificationSet))
		for i, n := range notificationSet {
			notificationIDs[i] = n.(string)
		}
		alertConfig.NotificationIDs = notificationIDs
	}

	// Serializar para JSON
	alertConfigJSON, err := json.Marshal(alertConfig)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configuração de alerta: %v", err))
	}

	// Criar configuração de alerta via API
	tflog.Debug(ctx, "Criando configuração de alerta")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/alert-configurations", alertConfigJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar configuração de alerta: %v", err))
	}

	// Deserializar resposta
	var createdAlertConfig AlertConfiguration
	if err := json.Unmarshal(resp, &createdAlertConfig); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdAlertConfig.ID == "" {
		return diag.FromErr(fmt.Errorf("configuração de alerta criada sem ID"))
	}

	d.SetId(createdAlertConfig.ID)
	return resourceAlertConfigurationRead(ctx, d, meta)
}

func resourceAlertConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da configuração de alerta não fornecido"))
	}

	// Buscar configuração de alerta via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo configuração de alerta com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/alert-configurations/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Configuração de alerta %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler configuração de alerta: %v", err))
	}

	// Deserializar resposta
	var alertConfig AlertConfiguration
	if err := json.Unmarshal(resp, &alertConfig); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	if err := d.Set("name", alertConfig.Name); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir name: %v", err))
	}

	if err := d.Set("description", alertConfig.Description); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir description: %v", err))
	}

	if err := d.Set("type", alertConfig.Type); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir type: %v", err))
	}

	if err := d.Set("enabled", alertConfig.Enabled); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir enabled: %v", err))
	}

	if err := d.Set("severity", alertConfig.Severity); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir severity: %v", err))
	}

	if err := d.Set("created", alertConfig.Created); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir created: %v", err))
	}

	if err := d.Set("updated", alertConfig.Updated); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir updated: %v", err))
	}

	// Converter mapa de condições para JSON
	if alertConfig.Conditions != nil {
		conditionsJSON, err := json.Marshal(alertConfig.Conditions)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao serializar condições: %v", err))
		}
		if err := d.Set("conditions", string(conditionsJSON)); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao definir conditions: %v", err))
		}
	}

	if alertConfig.Triggers != nil {
		if err := d.Set("triggers", alertConfig.Triggers); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao definir triggers: %v", err))
		}
	}

	if alertConfig.NotificationIDs != nil {
		if err := d.Set("notification_ids", alertConfig.NotificationIDs); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao definir notification_ids: %v", err))
		}
	}

	if alertConfig.OrgID != "" {
		if err := d.Set("org_id", alertConfig.OrgID); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao definir org_id: %v", err))
		}
	}

	return diags
}

func resourceAlertConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da configuração de alerta não fornecido"))
	}

	// Processar as condições (string JSON para map)
	var conditions map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("conditions").(string)), &conditions); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar condições: %v", err))
	}

	// Construir configuração de alerta atualizada
	alertConfig := &AlertConfiguration{
		ID:         id,
		Name:       d.Get("name").(string),
		Type:       d.Get("type").(string),
		Enabled:    d.Get("enabled").(bool),
		Conditions: conditions,
		Severity:   d.Get("severity").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		alertConfig.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		alertConfig.OrgID = v.(string)
	}

	// Processar triggers
	if v, ok := d.GetOk("triggers"); ok {
		triggerSet := v.(*schema.Set).List()
		triggers := make([]string, len(triggerSet))
		for i, t := range triggerSet {
			triggers[i] = t.(string)
		}
		alertConfig.Triggers = triggers
	}

	// Processar notification_ids
	if v, ok := d.GetOk("notification_ids"); ok {
		notificationSet := v.(*schema.Set).List()
		notificationIDs := make([]string, len(notificationSet))
		for i, n := range notificationSet {
			notificationIDs[i] = n.(string)
		}
		alertConfig.NotificationIDs = notificationIDs
	}

	// Serializar para JSON
	alertConfigJSON, err := json.Marshal(alertConfig)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configuração de alerta: %v", err))
	}

	// Atualizar configuração de alerta via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando configuração de alerta com ID: %s", id))
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/alert-configurations/%s", id), alertConfigJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar configuração de alerta: %v", err))
	}

	return resourceAlertConfigurationRead(ctx, d, meta)
}

func resourceAlertConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da configuração de alerta não fornecido"))
	}

	// Excluir configuração de alerta via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo configuração de alerta com ID: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/alert-configurations/%s", id), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao excluir configuração de alerta: %v", err))
	}

	d.SetId("")
	return diags
}
