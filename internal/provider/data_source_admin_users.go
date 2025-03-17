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

// AdminUsersResponse representa a resposta da API para a consulta de administradores
type AdminUsersResponse struct {
	Results     []AdminUser `json:"results"`
	TotalCount  int         `json:"totalCount"`
	NextPageURL string      `json:"nextPageUrl,omitempty"`
}

func dataSourceAdminUsers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAdminUsersRead,
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
			"sort": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"email", "firstName", "lastName", "created", "updated", "lastLogin"}, false),
							Description:  "Campo para ordenação",
						},
						"direction": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "asc",
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
				Description: "Número máximo de administradores a retornar",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambiente multi-tenant",
			},
			"admin_users": {
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
							Description: "Status do administrador",
						},
						"role_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "IDs dos papéis atribuídos ao administrador",
						},
						"is_mfa_enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Se a autenticação multifator está habilitada para o administrador",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data de criação do administrador",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data da última atualização do administrador",
						},
						"last_login": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data do último login do administrador",
						},
					},
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de administradores que correspondem aos filtros",
			},
		},
	}
}

func dataSourceAdminUsersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Construir URL base para a requisição
	url := "/api/v2/administrators"
	queryParams := ""

	// Aplicar filtros
	if filterList, ok := d.GetOk("filter"); ok && len(filterList.([]interface{})) > 0 {
		filter := filterList.([]interface{})[0].(map[string]interface{})

		if status, ok := filter["status"]; ok && status.(string) != "all" {
			queryParams += fmt.Sprintf("&status=%s", status.(string))
		}

		if search, ok := filter["search"]; ok && search.(string) != "" {
			queryParams += fmt.Sprintf("&search=%s", search.(string))
		}

		if hasRole, ok := filter["has_role"]; ok && hasRole.(string) != "" {
			queryParams += fmt.Sprintf("&hasRole=%s", hasRole.(string))
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
	tflog.Debug(ctx, fmt.Sprintf("Consultando administradores: %s%s", url, queryParams))
	resp, err := c.DoRequest(http.MethodGet, url+queryParams, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar administradores: %v", err))
	}

	// Deserializar resposta
	var response AdminUsersResponse
	if err := json.Unmarshal(resp, &response); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Converter lista de administradores para o formato do Terraform
	adminUsers := make([]map[string]interface{}, 0, len(response.Results))
	for _, admin := range response.Results {
		adminMap := map[string]interface{}{
			"id":             admin.ID,
			"email":          admin.Email,
			"first_name":     admin.FirstName,
			"last_name":      admin.LastName,
			"status":         admin.Status,
			"role_ids":       admin.RoleIDs,
			"is_mfa_enabled": admin.IsMFAEnabled,
			"created":        admin.Created,
			"updated":        admin.Updated,
			"last_login":     admin.LastLogin,
		}

		adminUsers = append(adminUsers, adminMap)
	}

	// Definir valores no state
	if err := d.Set("admin_users", adminUsers); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir admin_users no state: %v", err))
	}

	if err := d.Set("total_count", response.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir total_count no state: %v", err))
	}

	// Definir ID único para o data source (baseado no timestamp atual)
	d.SetId(fmt.Sprintf("admin-users-%d", time.Now().Unix()))

	return diags
}
