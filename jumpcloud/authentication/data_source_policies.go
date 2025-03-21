package authentication

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// AuthPoliciesResponse representa a resposta da API para listagem de políticas de autenticação
type AuthPoliciesResponse struct {
	Results     []AuthPolicy `json:"results"`
	TotalCount  int          `json:"totalCount"`
	NextPageURL string       `json:"nextPageUrl,omitempty"`
}

func DataSourcePolicies() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePoliciesRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Filtros para a listagem de políticas",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"sort": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Ordenação dos resultados (campo e direção, ex: {\"field\": \"name\", \"direction\": \"asc\"})",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     100,
				Description: "Número máximo de políticas a serem retornadas",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"auth_policies": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Lista de políticas de autenticação",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da política de autenticação",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome da política de autenticação",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Descrição da política de autenticação",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo da política (mfa, password, lockout, session)",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status da política (active, inactive, draft)",
						},
						"settings": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Configurações da política em formato JSON",
						},
						"priority": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Prioridade da política",
						},
						"apply_to_all_users": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Se a política se aplica a todos os usuários",
						},
						"effective_from": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data/hora de início de vigência da política",
						},
						"effective_until": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data/hora de término de vigência da política",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da organização",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data de criação da política",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data da última atualização da política",
						},
					},
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de políticas encontradas",
			},
		},
	}
}

func dataSourcePoliciesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir parâmetros de consulta
	queryParams := ""
	limit := d.Get("limit").(int)

	// Adicionar filtros
	if filters, ok := d.GetOk("filter"); ok {
		for k, v := range filters.(map[string]interface{}) {
			queryParams += fmt.Sprintf("%s=%s&", k, v.(string))
		}
	}

	// Adicionar organização
	if orgID, ok := d.GetOk("org_id"); ok {
		queryParams += fmt.Sprintf("orgId=%s&", orgID.(string))
	}

	// Adicionar ordenação
	if sort, ok := d.GetOk("sort"); ok {
		sortParams := sort.(map[string]interface{})
		if field, ok := sortParams["field"]; ok {
			queryParams += fmt.Sprintf("sort=%s&", field.(string))
			if direction, ok := sortParams["direction"]; ok {
				queryParams += fmt.Sprintf("direction=%s&", direction.(string))
			}
		}
	}

	// Adicionar limite
	queryParams += fmt.Sprintf("limit=%d", limit)

	// Remover o último & se existir
	if queryParams != "" {
		queryParams = "?" + queryParams
		if queryParams[len(queryParams)-1] == '&' {
			queryParams = queryParams[:len(queryParams)-1]
		}
	}

	// Consultar políticas via API
	tflog.Debug(ctx, "Consultando políticas de autenticação")
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/auth-policies%s", queryParams), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar políticas de autenticação: %v", err))
	}

	// Deserializar resposta
	var policiesResp AuthPoliciesResponse
	if err := json.Unmarshal(resp, &policiesResp); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	d.Set("total_count", policiesResp.TotalCount)

	// Preparar resultados
	policies := make([]map[string]interface{}, 0, len(policiesResp.Results))
	for _, policy := range policiesResp.Results {
		// Serializar settings para JSON
		settingsJSON, err := json.Marshal(policy.Settings)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao serializar settings: %v", err))
		}

		p := map[string]interface{}{
			"id":                 policy.ID,
			"name":               policy.Name,
			"description":        policy.Description,
			"type":               policy.Type,
			"status":             policy.Status,
			"settings":           string(settingsJSON),
			"priority":           policy.Priority,
			"apply_to_all_users": policy.ApplyToAllUsers,
			"effective_from":     policy.EffectiveFrom,
			"effective_until":    policy.EffectiveUntil,
			"created":            policy.Created,
			"updated":            policy.Updated,
		}

		if policy.OrgID != "" {
			p["org_id"] = policy.OrgID
		}

		policies = append(policies, p)
	}

	if err := d.Set("auth_policies", policies); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir auth_policies: %v", err))
	}

	// Definir ID do recurso de dados (timestamp para garantir unicidade)
	d.SetId(fmt.Sprintf("auth_policies_%d", time.Now().Unix()))

	return diags
}
