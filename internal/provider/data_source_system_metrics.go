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

// SystemMetric representa uma métrica de sistema no JumpCloud
type SystemMetric struct {
	SystemID   string                 `json:"systemId"`
	Hostname   string                 `json:"hostname,omitempty"`
	MetricName string                 `json:"metricName"`
	MetricType string                 `json:"metricType"` // cpu, memory, disk, network, etc.
	Value      float64                `json:"value"`
	Unit       string                 `json:"unit,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  string                 `json:"timestamp"`
}

// SystemMetricsResponse representa a resposta da API para consulta de métricas de sistema
type SystemMetricsResponse struct {
	Results     []SystemMetric `json:"results"`
	TotalCount  int            `json:"totalCount"`
	NextPageURL string         `json:"nextPageUrl,omitempty"`
}

func dataSourceSystemMetrics() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSystemMetricsRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"system_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar métricas por ID do sistema",
						},
						"hostname": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar métricas por hostname",
						},
						"metric_type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"cpu", "memory", "disk", "network", "process",
								"uptime", "load", "services", "all",
							}, false),
							Default:     "all",
							Description: "Tipo de métrica (cpu, memory, disk, etc.)",
						},
						"metric_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Nome específico da métrica",
						},
						"tags": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Filtrar por tags",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"start_time": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar métricas a partir desta data/hora (formato ISO8601)",
							Default:     time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
						},
						"end_time": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar métricas até esta data/hora (formato ISO8601)",
							Default:     time.Now().Format(time.RFC3339),
						},
					},
				},
			},
			"aggregation": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"function": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"avg", "min", "max", "sum", "count"}, false),
							Description:  "Função de agregação (avg, min, max, sum, count)",
						},
						"interval": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "1h",
							Description: "Intervalo de tempo para agregação (30s, 1m, 5m, 15m, 1h, etc.)",
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
							ValidateFunc: validation.StringInSlice([]string{"timestamp", "value", "metricName", "metricType", "hostname"}, false),
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
				Description: "Número máximo de métricas a serem retornadas",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambiente multi-tenant",
			},
			"metrics": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"system_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do sistema",
						},
						"hostname": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Hostname do sistema",
						},
						"metric_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome da métrica",
						},
						"metric_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo da métrica",
						},
						"value": {
							Type:        schema.TypeFloat,
							Computed:    true,
							Description: "Valor da métrica",
						},
						"unit": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Unidade de medida da métrica",
						},
						"tags": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Tags associadas à métrica",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"metadata": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Metadados adicionais da métrica em formato JSON",
						},
						"timestamp": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data/hora da coleta da métrica",
						},
					},
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de métricas que correspondem aos filtros",
			},
		},
	}
}

func dataSourceSystemMetricsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir URL base para a requisição
	url := "/api/v2/system-metrics"
	queryParams := ""

	// Aplicar filtros
	if filterList, ok := d.GetOk("filter"); ok && len(filterList.([]interface{})) > 0 {
		filter := filterList.([]interface{})[0].(map[string]interface{})

		if systemID, ok := filter["system_id"]; ok && systemID.(string) != "" {
			queryParams += fmt.Sprintf("&systemId=%s", systemID.(string))
		}

		if hostname, ok := filter["hostname"]; ok && hostname.(string) != "" {
			queryParams += fmt.Sprintf("&hostname=%s", hostname.(string))
		}

		if metricType, ok := filter["metric_type"]; ok && metricType.(string) != "all" {
			queryParams += fmt.Sprintf("&metricType=%s", metricType.(string))
		}

		if metricName, ok := filter["metric_name"]; ok && metricName.(string) != "" {
			queryParams += fmt.Sprintf("&metricName=%s", metricName.(string))
		}

		if startTime, ok := filter["start_time"]; ok && startTime.(string) != "" {
			queryParams += fmt.Sprintf("&startTime=%s", startTime.(string))
		}

		if endTime, ok := filter["end_time"]; ok && endTime.(string) != "" {
			queryParams += fmt.Sprintf("&endTime=%s", endTime.(string))
		}

		// Processar tags
		if tags, ok := filter["tags"]; ok {
			tagsList := tags.([]interface{})
			for _, tag := range tagsList {
				queryParams += fmt.Sprintf("&tags=%s", tag.(string))
			}
		}
	}

	// Aplicar agregação
	if aggregationList, ok := d.GetOk("aggregation"); ok && len(aggregationList.([]interface{})) > 0 {
		aggregation := aggregationList.([]interface{})[0].(map[string]interface{})
		function := aggregation["function"].(string)
		interval := aggregation["interval"].(string)

		queryParams += fmt.Sprintf("&aggregation=%s:%s", function, interval)
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
	tflog.Debug(ctx, fmt.Sprintf("Consultando métricas de sistema: %s%s", url, queryParams))
	resp, err := c.DoRequest(http.MethodGet, url+queryParams, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar métricas de sistema: %v", err))
	}

	// Deserializar resposta
	var response SystemMetricsResponse
	if err := json.Unmarshal(resp, &response); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Converter lista de métricas para o formato do Terraform
	metrics := make([]map[string]interface{}, 0, len(response.Results))
	for _, metric := range response.Results {
		metricMap := map[string]interface{}{
			"system_id":   metric.SystemID,
			"hostname":    metric.Hostname,
			"metric_name": metric.MetricName,
			"metric_type": metric.MetricType,
			"value":       metric.Value,
			"unit":        metric.Unit,
			"timestamp":   metric.Timestamp,
		}

		// Processar tags
		if metric.Tags != nil && len(metric.Tags) > 0 {
			metricMap["tags"] = metric.Tags
		}

		// Converter metadata para JSON
		if metric.Metadata != nil {
			metadataJSON, err := json.Marshal(metric.Metadata)
			if err != nil {
				return diag.FromErr(fmt.Errorf("erro ao serializar metadados: %v", err))
			}
			metricMap["metadata"] = string(metadataJSON)
		}

		metrics = append(metrics, metricMap)
	}

	// Definir valores no state
	if err := d.Set("metrics", metrics); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir metrics no state: %v", err))
	}

	if err := d.Set("total_count", response.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir total_count no state: %v", err))
	}

	// Definir ID único para o data source (baseado no timestamp atual)
	d.SetId(fmt.Sprintf("system-metrics-%d", time.Now().Unix()))

	return diags
}
