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

// AdminRolesResponse representa a resposta da API para a consulta de papéis de administrador
type AdminRolesResponse struct {
	Results     []AdminRole `json:"results"`
	TotalCount  int         `json:"totalCount"`
	NextPageURL string      `json:"nextPageUrl,omitempty"`
}

func dataSourceAdminRoles() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAdminRolesRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"system", "custom", "all"}, false),
							Description:  "Filtrar por tipo de papel (system, custom, all)",
							Default:      "all",
						},
						"scope": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"global", "org", "resource", "all"}, false),
							Description:  "Filtrar por escopo do papel (global, org, resource, all)",
							Default:      "all",
						},
						"search": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Termo de busca para filtrar papéis (nome, descrição)",
						},
						"has_permission": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar papéis que possuem uma permissão específica",
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
							ValidateFunc: validation.StringInSlice([]string{"name", "type", "scope", "created", "updated"}, false),
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
				Description: "Número máximo de papéis a retornar",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambiente multi-tenant",
			},
			"admin_roles": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do papel de administrador",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome do papel de administrador",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Descrição do papel de administrador",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo do papel (system, custom)",
						},
						"scope": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Escopo do papel (global, org, resource)",
						},
						"permissions": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Lista de permissões do papel",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data de criação do papel",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data da última atualização do papel",
						},
					},
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de papéis que correspondem aos filtros",
			},
		},
	}
}

func dataSourceAdminRolesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Construir URL base para a requisição
	url := "/api/v2/admin-roles"
	queryParams := ""

	// Aplicar filtros
	if filterList, ok := d.GetOk("filter"); ok && len(filterList.([]interface{})) > 0 {
		filter := filterList.([]interface{})[0].(map[string]interface{})

		if roleType, ok := filter["type"]; ok && roleType.(string) != "all" {
			queryParams += fmt.Sprintf("&type=%s", roleType.(string))
		}

		if scope, ok := filter["scope"]; ok && scope.(string) != "all" {
			queryParams += fmt.Sprintf("&scope=%s", scope.(string))
		}

		if search, ok := filter["search"]; ok && search.(string) != "" {
			queryParams += fmt.Sprintf("&search=%s", search.(string))
		}

		if hasPermission, ok := filter["has_permission"]; ok && hasPermission.(string) != "" {
			queryParams += fmt.Sprintf("&hasPermission=%s", hasPermission.(string))
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
	tflog.Debug(ctx, fmt.Sprintf("Consultando papéis de administrador: %s%s", url, queryParams))
	resp, err := c.DoRequest(http.MethodGet, url+queryParams, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar papéis de administrador: %v", err))
	}

	// Deserializar resposta
	var response AdminRolesResponse
	if err := json.Unmarshal(resp, &response); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Converter lista de papéis para o formato do Terraform
	adminRoles := make([]map[string]interface{}, 0, len(response.Results))
	for _, role := range response.Results {
		roleMap := map[string]interface{}{
			"id":          role.ID,
			"name":        role.Name,
			"description": role.Description,
			"type":        role.Type,
			"scope":       role.Scope,
			"permissions": role.Permissions,
			"created":     role.Created,
			"updated":     role.Updated,
		}

		adminRoles = append(adminRoles, roleMap)
	}

	// Definir valores no state
	if err := d.Set("admin_roles", adminRoles); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir admin_roles no state: %v", err))
	}

	if err := d.Set("total_count", response.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir total_count no state: %v", err))
	}

	// Definir ID único para o data source (baseado no timestamp atual)
	d.SetId(fmt.Sprintf("admin-roles-%d", time.Now().Unix()))

	return diags
}
