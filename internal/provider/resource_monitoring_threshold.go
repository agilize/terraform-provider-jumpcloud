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

// MonitoringThreshold representa um limiar de monitoramento no JumpCloud
type MonitoringThreshold struct {
	ID           string                 `json:"_id,omitempty"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	MetricType   string                 `json:"metricType"` // cpu, memory, disk, network, etc.
	ResourceType string                 `json:"resourceType"`
	Operator     string                 `json:"operator"`  // gt, lt, eq, ne, etc.
	Threshold    float64                `json:"threshold"` // valor numérico do limiar
	Duration     int                    `json:"duration"`  // duração em segundos
	Severity     string                 `json:"severity,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Actions      map[string]interface{} `json:"actions,omitempty"`
	OrgID        string                 `json:"orgId,omitempty"`
	Created      string                 `json:"created,omitempty"`
	Updated      string                 `json:"updated,omitempty"`
}

func resourceMonitoringThreshold() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMonitoringThresholdCreate,
		ReadContext:   resourceMonitoringThresholdRead,
		UpdateContext: resourceMonitoringThresholdUpdate,
		DeleteContext: resourceMonitoringThresholdDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome do limiar de monitoramento",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição do limiar de monitoramento",
			},
			"metric_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"cpu", "memory", "disk", "network", "application", "process",
					"login", "security", "system_uptime", "agent", "services",
				}, false),
				Description: "Tipo de métrica monitorada (cpu, memory, disk, etc.)",
			},
			"resource_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"system", "user", "group", "application", "directory", "policy",
					"organization", "device", "service",
				}, false),
				Description: "Tipo de recurso monitorado",
			},
			"operator": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"gt", "lt", "eq", "ne", "ge", "le",
				}, false),
				Description: "Operador de comparação (gt=maior que, lt=menor que, eq=igual, ne=diferente, etc.)",
			},
			"threshold": {
				Type:        schema.TypeFloat,
				Required:    true,
				Description: "Valor numérico do limiar",
			},
			"duration": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Duração em segundos para considerar o limiar atingido",
			},
			"severity": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "medium",
				ValidateFunc: validation.StringInSlice([]string{"critical", "high", "medium", "low", "info"}, false),
				Description:  "Severidade do alerta quando o limiar é atingido",
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Tags associadas ao limiar",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"actions": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Ações a serem executadas quando o limiar é atingido, em formato JSON",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					jsonStr := val.(string)
					var js map[string]interface{}
					if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
						errs = append(errs, fmt.Errorf("%q: JSON inválido: %s", key, err))
					}
					return
				},
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação do limiar",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização do limiar",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceMonitoringThresholdCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Processar ações (string JSON para map), se fornecido
	var actions map[string]interface{}
	if v, ok := d.GetOk("actions"); ok {
		if err := json.Unmarshal([]byte(v.(string)), &actions); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar ações: %v", err))
		}
	}

	// Construir limiar de monitoramento
	threshold := &MonitoringThreshold{
		Name:         d.Get("name").(string),
		MetricType:   d.Get("metric_type").(string),
		ResourceType: d.Get("resource_type").(string),
		Operator:     d.Get("operator").(string),
		Threshold:    d.Get("threshold").(float64),
		Duration:     d.Get("duration").(int),
		Severity:     d.Get("severity").(string),
		Actions:      actions,
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		threshold.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		threshold.OrgID = v.(string)
	}

	// Processar tags
	if v, ok := d.GetOk("tags"); ok {
		tagsSet := v.(*schema.Set).List()
		tags := make([]string, len(tagsSet))
		for i, t := range tagsSet {
			tags[i] = t.(string)
		}
		threshold.Tags = tags
	}

	// Serializar para JSON
	thresholdJSON, err := json.Marshal(threshold)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar limiar de monitoramento: %v", err))
	}

	// Criar limiar de monitoramento via API
	tflog.Debug(ctx, "Criando limiar de monitoramento")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/monitoring-thresholds", thresholdJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar limiar de monitoramento: %v", err))
	}

	// Deserializar resposta
	var createdThreshold MonitoringThreshold
	if err := json.Unmarshal(resp, &createdThreshold); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdThreshold.ID == "" {
		return diag.FromErr(fmt.Errorf("limiar de monitoramento criado sem ID"))
	}

	d.SetId(createdThreshold.ID)
	return resourceMonitoringThresholdRead(ctx, d, meta)
}

func resourceMonitoringThresholdRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do limiar de monitoramento não fornecido"))
	}

	// Buscar limiar de monitoramento via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo limiar de monitoramento com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/monitoring-thresholds/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Limiar de monitoramento %s não encontrado, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler limiar de monitoramento: %v", err))
	}

	// Deserializar resposta
	var threshold MonitoringThreshold
	if err := json.Unmarshal(resp, &threshold); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("name", threshold.Name)
	d.Set("description", threshold.Description)
	d.Set("metric_type", threshold.MetricType)
	d.Set("resource_type", threshold.ResourceType)
	d.Set("operator", threshold.Operator)
	d.Set("threshold", threshold.Threshold)
	d.Set("duration", threshold.Duration)
	d.Set("severity", threshold.Severity)
	d.Set("created", threshold.Created)
	d.Set("updated", threshold.Updated)

	// Converter actions para JSON
	if threshold.Actions != nil {
		actionsJSON, err := json.Marshal(threshold.Actions)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao serializar actions: %v", err))
		}
		d.Set("actions", string(actionsJSON))
	}

	if threshold.Tags != nil {
		d.Set("tags", threshold.Tags)
	}

	if threshold.OrgID != "" {
		d.Set("org_id", threshold.OrgID)
	}

	return diags
}

func resourceMonitoringThresholdUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do limiar de monitoramento não fornecido"))
	}

	// Processar ações (string JSON para map), se fornecido
	var actions map[string]interface{}
	if v, ok := d.GetOk("actions"); ok {
		if err := json.Unmarshal([]byte(v.(string)), &actions); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar ações: %v", err))
		}
	}

	// Construir limiar de monitoramento atualizado
	threshold := &MonitoringThreshold{
		ID:           id,
		Name:         d.Get("name").(string),
		MetricType:   d.Get("metric_type").(string),
		ResourceType: d.Get("resource_type").(string),
		Operator:     d.Get("operator").(string),
		Threshold:    d.Get("threshold").(float64),
		Duration:     d.Get("duration").(int),
		Severity:     d.Get("severity").(string),
		Actions:      actions,
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		threshold.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		threshold.OrgID = v.(string)
	}

	// Processar tags
	if v, ok := d.GetOk("tags"); ok {
		tagsSet := v.(*schema.Set).List()
		tags := make([]string, len(tagsSet))
		for i, t := range tagsSet {
			tags[i] = t.(string)
		}
		threshold.Tags = tags
	}

	// Serializar para JSON
	thresholdJSON, err := json.Marshal(threshold)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar limiar de monitoramento: %v", err))
	}

	// Atualizar limiar de monitoramento via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando limiar de monitoramento: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/monitoring-thresholds/%s", id), thresholdJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar limiar de monitoramento: %v", err))
	}

	// Deserializar resposta
	var updatedThreshold MonitoringThreshold
	if err := json.Unmarshal(resp, &updatedThreshold); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceMonitoringThresholdRead(ctx, d, meta)
}

func resourceMonitoringThresholdDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do limiar de monitoramento não fornecido"))
	}

	// Excluir limiar de monitoramento via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo limiar de monitoramento: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/monitoring-thresholds/%s", id), nil)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Limiar de monitoramento %s não encontrado, considerando excluído", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir limiar de monitoramento: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
