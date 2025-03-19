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

// AdminAuditLogEntry representa uma entrada de log de auditoria de administrador
type AdminAuditLogEntry struct {
	ID           string                 `json:"_id,omitempty"`
	AdminUserID  string                 `json:"adminUserId,omitempty"`
	AdminEmail   string                 `json:"adminEmail,omitempty"`
	Action       string                 `json:"action,omitempty"`
	ResourceType string                 `json:"resourceType,omitempty"`
	ResourceID   string                 `json:"resourceId,omitempty"`
	ResourceName string                 `json:"resourceName,omitempty"`
	Changes      map[string]interface{} `json:"changes,omitempty"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"errorMessage,omitempty"`
	ClientIP     string                 `json:"clientIp,omitempty"`
	UserAgent    string                 `json:"userAgent,omitempty"`
	OrgID        string                 `json:"orgId,omitempty"`
	Timestamp    string                 `json:"timestamp,omitempty"`
	OperationID  string                 `json:"operationId,omitempty"`
}

// AdminAuditLogsResponse representa a resposta da API para a consulta de logs de auditoria
type AdminAuditLogsResponse struct {
	Results     []AdminAuditLogEntry `json:"results"`
	TotalCount  int                  `json:"totalCount"`
	NextPageURL string               `json:"nextPageUrl,omitempty"`
}

func dataSourceAdminAuditLogs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAdminAuditLogsRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"admin_user_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar por ID do administrador",
						},
						"admin_email": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar por email do administrador",
						},
						"action": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"create", "read", "update", "delete", "login", "all"}, false),
							Description:  "Filtrar por tipo de ação (create, read, update, delete, login, all)",
							Default:      "all",
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
							Description: "Data/hora de início para filtro (formato ISO8601)",
						},
						"end_time": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Data/hora de fim para filtro (formato ISO8601)",
						},
						"success": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Filtrar por resultado da operação (sucesso/falha)",
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
							ValidateFunc: validation.StringInSlice([]string{"timestamp", "adminEmail", "action", "resourceType"}, false),
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
				Description: "Número máximo de registros de auditoria a retornar",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambiente multi-tenant",
			},
			"audit_logs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do registro de auditoria",
						},
						"admin_user_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do administrador",
						},
						"admin_email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Email do administrador",
						},
						"action": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo de ação realizada",
						},
						"resource_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo do recurso afetado",
						},
						"resource_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do recurso afetado",
						},
						"resource_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome do recurso afetado",
						},
						"changes": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Alterações realizadas (em formato JSON)",
						},
						"success": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Se a operação foi bem-sucedida",
						},
						"error_message": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Mensagem de erro (se houver)",
						},
						"client_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Endereço IP do cliente",
						},
						"user_agent": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "User-Agent do cliente",
						},
						"timestamp": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data/hora da ação",
						},
						"operation_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da operação (útil para rastrear operações relacionadas)",
						},
					},
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de registros de auditoria que correspondem aos filtros",
			},
		},
	}
}

func dataSourceAdminAuditLogsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir URL base para a requisição
	url := "/api/v2/admin-audit-logs"
	queryParams := ""

	// Aplicar filtros
	if filterList, ok := d.GetOk("filter"); ok && len(filterList.([]interface{})) > 0 {
		filter := filterList.([]interface{})[0].(map[string]interface{})

		if adminUserID, ok := filter["admin_user_id"]; ok && adminUserID.(string) != "" {
			queryParams += fmt.Sprintf("&adminUserId=%s", adminUserID.(string))
		}

		if adminEmail, ok := filter["admin_email"]; ok && adminEmail.(string) != "" {
			queryParams += fmt.Sprintf("&adminEmail=%s", adminEmail.(string))
		}

		if action, ok := filter["action"]; ok && action.(string) != "all" {
			queryParams += fmt.Sprintf("&action=%s", action.(string))
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

		if success, ok := filter["success"]; ok {
			queryParams += fmt.Sprintf("&success=%t", success.(bool))
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
	tflog.Debug(ctx, fmt.Sprintf("Consultando logs de auditoria de administradores: %s%s", url, queryParams))
	resp, err := c.DoRequest(http.MethodGet, url+queryParams, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar logs de auditoria: %v", err))
	}

	// Deserializar resposta
	var response AdminAuditLogsResponse
	if err := json.Unmarshal(resp, &response); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Converter lista de logs para o formato do Terraform
	auditLogs := make([]map[string]interface{}, 0, len(response.Results))
	for _, log := range response.Results {
		// Converter o mapa de alterações para JSON
		var changesJSON string
		if log.Changes != nil {
			changesBytes, err := json.Marshal(log.Changes)
			if err != nil {
				return diag.FromErr(fmt.Errorf("erro ao serializar alterações: %v", err))
			}
			changesJSON = string(changesBytes)
		}

		logMap := map[string]interface{}{
			"id":            log.ID,
			"admin_user_id": log.AdminUserID,
			"admin_email":   log.AdminEmail,
			"action":        log.Action,
			"resource_type": log.ResourceType,
			"resource_id":   log.ResourceID,
			"resource_name": log.ResourceName,
			"changes":       changesJSON,
			"success":       log.Success,
			"error_message": log.ErrorMessage,
			"client_ip":     log.ClientIP,
			"user_agent":    log.UserAgent,
			"timestamp":     log.Timestamp,
			"operation_id":  log.OperationID,
		}

		auditLogs = append(auditLogs, logMap)
	}

	// Definir valores no state
	if err := d.Set("audit_logs", auditLogs); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir audit_logs no state: %v", err))
	}

	if err := d.Set("total_count", response.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir total_count no state: %v", err))
	}

	// Definir ID único para o data source (baseado no timestamp atual)
	d.SetId(fmt.Sprintf("admin-audit-logs-%d", time.Now().Unix()))

	return diags
}
