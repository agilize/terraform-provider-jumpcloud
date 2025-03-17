package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Alert representa um alerta no JumpCloud
type Alert struct {
	ID             string                 `json:"_id,omitempty"`
	ConfigID       string                 `json:"configId,omitempty"`
	ConfigName     string                 `json:"configName,omitempty"`
	Type           string                 `json:"type,omitempty"`
	Status         string                 `json:"status,omitempty"` // active, resolved, acknowledged
	ResourceType   string                 `json:"resourceType,omitempty"`
	ResourceID     string                 `json:"resourceId,omitempty"`
	ResourceName   string                 `json:"resourceName,omitempty"`
	Message        string                 `json:"message,omitempty"`
	Severity       string                 `json:"severity,omitempty"`
	Data           map[string]interface{} `json:"data,omitempty"`
	TriggeredBy    string                 `json:"triggeredBy,omitempty"`
	ResolvedBy     string                 `json:"resolvedBy,omitempty"`
	AcknowledgedBy string                 `json:"acknowledgedBy,omitempty"`
	Created        string                 `json:"created,omitempty"`
	Updated        string                 `json:"updated,omitempty"`
	OrgID          string                 `json:"orgId,omitempty"`
}

// AlertsResponse representa a resposta da API para listagem de alertas
type AlertsResponse struct {
	Results     []Alert `json:"results"`
	TotalCount  int     `json:"totalCount"`
	NextPageURL string  `json:"nextPageUrl,omitempty"`
}

func dataSourceAlerts() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAlertsRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"active", "resolved", "acknowledged", "all"}, false),
							Description:  "Status dos alertas a serem retornados (active, resolved, acknowledged, all)",
							Default:      "active",
						},
						"severity": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"critical", "high", "medium", "low", "info", "all"}, false),
							Description:  "Severidade dos alertas a serem retornados",
							Default:      "all",
						},
						"type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Tipo de alerta a ser filtrado",
						},
						"resource_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar por tipo de recurso",
						},
						"resource_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar por ID do recurso",
						},
						"start_time": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar alertas a partir desta data/hora (formato ISO8601)",
						},
						"end_time": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar alertas até esta data/hora (formato ISO8601)",
						},
						"config_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar alertas de uma configuração específica",
						},
					},
				},
			},
			"sort": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"created", "updated", "severity", "status", "type"}, false),
							Description:  "Campo para ordenação",
						},
						"direction": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "desc",
							ValidateFunc: validation.StringInSlice([]string{"asc", "desc"}, false),
							Description:  "Direção da ordenação (asc, desc)",
						},
					},
				},
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     100,
				Description: "Número máximo de alertas a serem retornados",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambiente multi-tenant",
			},
			"alerts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do alerta",
						},
						"config_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da configuração que gerou o alerta",
						},
						"config_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome da configuração que gerou o alerta",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo do alerta",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status do alerta (active, resolved, acknowledged)",
						},
						"resource_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo do recurso que gerou o alerta",
						},
						"resource_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do recurso que gerou o alerta",
						},
						"resource_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome do recurso que gerou o alerta",
						},
						"message": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Mensagem do alerta",
						},
						"severity": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Severidade do alerta",
						},
						"data": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Dados detalhados do alerta em formato JSON",
						},
						"triggered_by": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do usuário ou sistema que acionou o alerta",
						},
						"resolved_by": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do usuário que resolveu o alerta",
						},
						"acknowledged_by": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do usuário que reconheceu o alerta",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data de criação do alerta",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data da última atualização do alerta",
						},
					},
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de alertas que correspondem aos filtros",
			},
		},
	}
}

func dataSourceAlertsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Construir URL base para a requisição
	url := "/api/v2/alerts"
	queryParams := ""

	// Aplicar filtros
	if filterList, ok := d.GetOk("filter"); ok && len(filterList.([]interface{})) > 0 {
		filter := filterList.([]interface{})[0].(map[string]interface{})

		if status, ok := filter["status"]; ok && status.(string) != "all" {
			queryParams += fmt.Sprintf("&status=%s", status.(string))
		}

		if severity, ok := filter["severity"]; ok && severity.(string) != "all" {
			queryParams += fmt.Sprintf("&severity=%s", severity.(string))
		}

		if alertType, ok := filter["type"]; ok && alertType.(string) != "" {
			queryParams += fmt.Sprintf("&type=%s", alertType.(string))
		}

		if resourceType, ok := filter["resource_type"]; ok && resourceType.(string) != "" {
			queryParams += fmt.Sprintf("&resourceType=%s", resourceType.(string))
		}

		if resourceID, ok := filter["resource_id"]; ok && resourceID.(string) != "" {
			queryParams += fmt.Sprintf("&resourceId=%s", resourceID.(string))
		}

		if startTime, ok := filter["start_time"]; ok && startTime.(string) != "" {
			queryParams += fmt.Sprintf("&startTime=%s", startTime.(string))
		}

		if endTime, ok := filter["end_time"]; ok && endTime.(string) != "" {
			queryParams += fmt.Sprintf("&endTime=%s", endTime.(string))
		}

		if configID, ok := filter["config_id"]; ok && configID.(string) != "" {
			queryParams += fmt.Sprintf("&configId=%s", configID.(string))
		}
	}

	// Aplicar ordenação
	if sortList, ok := d.GetOk("sort"); ok && len(sortList.([]interface{})) > 0 {
		sort := sortList.([]interface{})[0].(map[string]interface{})
		field := sort["field"].(string)
		direction := sort["direction"].(string)

		queryParams += fmt.Sprintf("&sort=%s:%s", field, direction)
	}

	// Aplicar limite
	limit := d.Get("limit").(int)
	queryParams += fmt.Sprintf("&limit=%d", limit)

	// Adicionar organizationID se fornecido
	if orgID, ok := d.GetOk("org_id"); ok {
		queryParams += fmt.Sprintf("&orgId=%s", orgID.(string))
	}

	// Remover o '&' inicial se existir
	if len(queryParams) > 0 {
		queryParams = "?" + queryParams[1:]
	}

	// Fazer a requisição à API
	tflog.Debug(ctx, fmt.Sprintf("Consultando alertas: %s%s", url, queryParams))
	resp, err := c.DoRequest(http.MethodGet, url+queryParams, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar alertas: %v", err))
	}

	// Deserializar resposta
	var response AlertsResponse
	if err := json.Unmarshal(resp, &response); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Converter lista de alertas para o formato do Terraform
	alerts := make([]map[string]interface{}, 0, len(response.Results))
	for _, alert := range response.Results {
		alertMap := map[string]interface{}{
			"id":              alert.ID,
			"config_id":       alert.ConfigID,
			"config_name":     alert.ConfigName,
			"type":            alert.Type,
			"status":          alert.Status,
			"resource_type":   alert.ResourceType,
			"resource_id":     alert.ResourceID,
			"resource_name":   alert.ResourceName,
			"message":         alert.Message,
			"severity":        alert.Severity,
			"triggered_by":    alert.TriggeredBy,
			"resolved_by":     alert.ResolvedBy,
			"acknowledged_by": alert.AcknowledgedBy,
			"created":         alert.Created,
			"updated":         alert.Updated,
		}

		// Converter dados para formato JSON
		if alert.Data != nil {
			dataJSON, err := json.Marshal(alert.Data)
			if err != nil {
				return diag.FromErr(fmt.Errorf("erro ao serializar dados do alerta: %v", err))
			}
			alertMap["data"] = string(dataJSON)
		}

		alerts = append(alerts, alertMap)
	}

	// Definir valores no state
	if err := d.Set("alerts", alerts); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir alerts no state: %v", err))
	}

	if err := d.Set("total_count", response.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir total_count no state: %v", err))
	}

	// Definir ID único para o data source (baseado no timestamp atual)
	d.SetId(fmt.Sprintf("alerts-%d", time.Now().Unix()))

	return diags
}
