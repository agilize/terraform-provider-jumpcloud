package admin_roles

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

func DataSourceAuditLogs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAuditLogsRead,
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
							Description: "Filtrar logs por ID do administrador",
						},
						"admin_email": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar logs por email do administrador",
						},
						"action": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"create", "read", "update", "delete", "list", "login", "all"}, false),
							Description:  "Filtrar logs por tipo de ação (create, read, update, delete, list, login, all)",
							Default:      "all",
						},
						"resource_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar logs por tipo de recurso",
						},
						"resource_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar logs por ID do recurso",
						},
						"success": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Filtrar logs por status de sucesso",
						},
						"start_time": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar logs a partir de uma data/hora (RFC3339)",
						},
						"end_time": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar logs até uma data/hora (RFC3339)",
						},
						"search": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Termo de busca para filtrar logs",
						},
					},
				},
			},
			"logs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do log de auditoria",
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
							Description: "Tipo do recurso",
						},
						"resource_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do recurso",
						},
						"resource_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome do recurso",
						},
						"changes": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Alterações realizadas (JSON)",
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
							Description: "IP do cliente",
						},
						"user_agent": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "User-Agent do cliente",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da organização",
						},
						"timestamp": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data/hora do evento",
						},
						"operation_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da operação",
						},
					},
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de logs de auditoria encontrados",
			},
		},
	}
}

func dataSourceAuditLogsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir parâmetros de consulta
	queryParams := ""

	if filterList, ok := d.GetOk("filter"); ok && len(filterList.([]interface{})) > 0 {
		filter := filterList.([]interface{})[0].(map[string]interface{})

		// Admin User ID
		if adminUserID, ok := filter["admin_user_id"].(string); ok && adminUserID != "" {
			queryParams += fmt.Sprintf("adminUserId=%s&", adminUserID)
		}

		// Admin Email
		if adminEmail, ok := filter["admin_email"].(string); ok && adminEmail != "" {
			queryParams += fmt.Sprintf("adminEmail=%s&", adminEmail)
		}

		// Action
		if action, ok := filter["action"].(string); ok && action != "all" {
			queryParams += fmt.Sprintf("action=%s&", action)
		}

		// Resource Type
		if resourceType, ok := filter["resource_type"].(string); ok && resourceType != "" {
			queryParams += fmt.Sprintf("resourceType=%s&", resourceType)
		}

		// Resource ID
		if resourceID, ok := filter["resource_id"].(string); ok && resourceID != "" {
			queryParams += fmt.Sprintf("resourceId=%s&", resourceID)
		}

		// Success
		if success, ok := filter["success"].(bool); ok {
			queryParams += fmt.Sprintf("success=%t&", success)
		}

		// Start Time
		if startTime, ok := filter["start_time"].(string); ok && startTime != "" {
			queryParams += fmt.Sprintf("startTime=%s&", startTime)
		}

		// End Time
		if endTime, ok := filter["end_time"].(string); ok && endTime != "" {
			queryParams += fmt.Sprintf("endTime=%s&", endTime)
		}

		// Search
		if search, ok := filter["search"].(string); ok && search != "" {
			queryParams += fmt.Sprintf("search=%s&", search)
		}
	}

	// Remover o último & se existir
	if queryParams != "" {
		queryParams = "?" + queryParams
		if queryParams[len(queryParams)-1] == '&' {
			queryParams = queryParams[:len(queryParams)-1]
		}
	}

	// Consultar logs de auditoria via API
	tflog.Debug(ctx, "Consultando logs de auditoria de administradores")
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/admin-audit-logs%s", queryParams), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar logs de auditoria: %v", err))
	}

	// Deserializar resposta
	var auditLogsResp AdminAuditLogsResponse
	if err := json.Unmarshal(resp, &auditLogsResp); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	d.Set("total_count", auditLogsResp.TotalCount)

	// Preparar resultados
	logs := make([]map[string]interface{}, 0, len(auditLogsResp.Results))
	for _, entry := range auditLogsResp.Results {
		log := map[string]interface{}{
			"id":            entry.ID,
			"admin_user_id": entry.AdminUserID,
			"admin_email":   entry.AdminEmail,
			"action":        entry.Action,
			"resource_type": entry.ResourceType,
			"resource_id":   entry.ResourceID,
			"resource_name": entry.ResourceName,
			"success":       entry.Success,
			"error_message": entry.ErrorMessage,
			"client_ip":     entry.ClientIP,
			"user_agent":    entry.UserAgent,
			"timestamp":     entry.Timestamp,
			"operation_id":  entry.OperationID,
		}

		// Campos opcionais
		if entry.OrgID != "" {
			log["org_id"] = entry.OrgID
		}

		// Serializar changes para JSON
		if entry.Changes != nil {
			changesJSON, err := json.Marshal(entry.Changes)
			if err != nil {
				tflog.Warn(ctx, fmt.Sprintf("Erro ao serializar changes para o log %s: %v", entry.ID, err))
			} else {
				log["changes"] = string(changesJSON)
			}
		}

		logs = append(logs, log)
	}

	if err := d.Set("logs", logs); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir logs: %v", err))
	}

	// Definir ID do recurso de dados (timestamp para garantir unicidade)
	d.SetId(fmt.Sprintf("admin_audit_logs_%d", time.Now().Unix()))

	return diags
}
