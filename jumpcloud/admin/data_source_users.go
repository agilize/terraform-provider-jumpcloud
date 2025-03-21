package admin

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

// AdminUsersResponse representa a resposta da API para a consulta de administradores
type AdminUsersResponse struct {
	Results     []AdminUser `json:"results"`
	TotalCount  int         `json:"totalCount"`
	NextPageURL string      `json:"nextPageUrl,omitempty"`
}

func DataSourceUsers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUsersRead,
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
							ValidateFunc: validation.StringInSlice([]string{"active", "pending", "disabled", "all"}, false),
							Description:  "Filtrar por status do administrador (active, pending, disabled, all)",
							Default:      "all",
						},
						"search": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Termo de busca para filtrar administradores (email, nome)",
						},
						"has_role": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar administradores que possuem um papel específico (ID do papel)",
						},
					},
				},
			},
			"users": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do administrador",
						},
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Email do administrador",
						},
						"first_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Primeiro nome do administrador",
						},
						"last_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Sobrenome do administrador",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status do administrador (active, pending, disabled)",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da organização",
						},
						"is_mfa_enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Se o MFA está habilitado para o administrador",
						},
						"role_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "IDs dos papéis atribuídos ao administrador",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data de criação do administrador",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data de última atualização do administrador",
						},
						"last_login": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data do último login do administrador",
						},
					},
				},
			},
		},
	}
}

func dataSourceUsersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir parâmetros de consulta
	queryParams := ""

	if filterList, ok := d.GetOk("filter"); ok && len(filterList.([]interface{})) > 0 {
		filter := filterList.([]interface{})[0].(map[string]interface{})

		// Status
		if status, ok := filter["status"].(string); ok && status != "all" {
			queryParams += fmt.Sprintf("status=%s&", status)
		}

		// Search
		if search, ok := filter["search"].(string); ok && search != "" {
			queryParams += fmt.Sprintf("search=%s&", search)
		}

		// Has Role
		if roleID, ok := filter["has_role"].(string); ok && roleID != "" {
			queryParams += fmt.Sprintf("hasRole=%s&", roleID)
		}
	}

	// Remover o último & se existir
	if queryParams != "" {
		queryParams = "?" + queryParams
		if queryParams[len(queryParams)-1] == '&' {
			queryParams = queryParams[:len(queryParams)-1]
		}
	}

	// Consultar administradores via API
	tflog.Debug(ctx, "Consultando administradores")
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/administrators%s", queryParams), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar administradores: %v", err))
	}

	// Deserializar resposta
	var adminUsersResp AdminUsersResponse
	if err := json.Unmarshal(resp, &adminUsersResp); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Preparar resultados
	users := make([]map[string]interface{}, 0, len(adminUsersResp.Results))
	for _, admin := range adminUsersResp.Results {
		user := map[string]interface{}{
			"id":             admin.ID,
			"email":          admin.Email,
			"first_name":     admin.FirstName,
			"last_name":      admin.LastName,
			"status":         admin.Status,
			"is_mfa_enabled": admin.IsMFAEnabled,
			"created":        admin.Created,
			"updated":        admin.Updated,
			"last_login":     admin.LastLogin,
		}

		// Campos opcionais
		if admin.OrgID != "" {
			user["org_id"] = admin.OrgID
		}

		// Role IDs
		if admin.RoleIDs != nil {
			user["role_ids"] = admin.RoleIDs
		}

		users = append(users, user)
	}

	if err := d.Set("users", users); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir users: %v", err))
	}

	// Definir ID do recurso de dados (timestamp para garantir unicidade)
	d.SetId(fmt.Sprintf("admin_users_%d", time.Now().Unix()))

	return diags
}
