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
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// AdminRolesResponse representa a resposta da API para a consulta de papéis de administrador
type AdminRolesResponse struct {
	Results     []AdminRole `json:"results"`
	TotalCount  int         `json:"totalCount"`
	NextPageURL string      `json:"nextPageUrl,omitempty"`
}

func DataSourceRoles() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRolesRead,
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
					},
				},
			},
			"roles": {
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
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da organização",
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
							Description: "Data de última atualização do papel",
						},
					},
				},
			},
		},
	}
}

func dataSourceRolesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir parâmetros de consulta
	queryParams := ""

	if filterList, ok := d.GetOk("filter"); ok && len(filterList.([]interface{})) > 0 {
		filter := filterList.([]interface{})[0].(map[string]interface{})

		// Type
		if roleType, ok := filter["type"].(string); ok && roleType != "all" {
			queryParams += fmt.Sprintf("type=%s&", roleType)
		}

		// Scope
		if scope, ok := filter["scope"].(string); ok && scope != "all" {
			queryParams += fmt.Sprintf("scope=%s&", scope)
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

	// Consultar papéis de administrador via API
	tflog.Debug(ctx, "Consultando papéis de administrador")
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/admin-roles%s", queryParams), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar papéis de administrador: %v", err))
	}

	// Deserializar resposta
	var adminRolesResp AdminRolesResponse
	if err := json.Unmarshal(resp, &adminRolesResp); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Preparar resultados
	roles := make([]map[string]interface{}, 0, len(adminRolesResp.Results))
	for _, role := range adminRolesResp.Results {
		roleMap := map[string]interface{}{
			"id":          role.ID,
			"name":        role.Name,
			"description": role.Description,
			"type":        role.Type,
			"scope":       role.Scope,
			"created":     role.Created,
			"updated":     role.Updated,
		}

		// Campos opcionais
		if role.OrgID != "" {
			roleMap["org_id"] = role.OrgID
		}

		// Permissions
		if role.Permissions != nil {
			roleMap["permissions"] = role.Permissions
		}

		roles = append(roles, roleMap)
	}

	if err := d.Set("roles", roles); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir roles: %v", err))
	}

	// Definir ID do recurso de dados (timestamp para garantir unicidade)
	d.SetId(fmt.Sprintf("admin_roles_%d", time.Now().Unix()))

	return diags
}
