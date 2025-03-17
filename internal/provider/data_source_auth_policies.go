package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

func dataSourceAuthPolicies() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAuthPoliciesRead,
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

func dataSourceAuthPoliciesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir a URL com query params
	url := "/api/v2/auth-policies"
	queryParams := ""

	// Adicionar filtros
	if v, ok := d.GetOk("filter"); ok {
		filters := v.(map[string]interface{})
		for key, value := range filters {
			if queryParams == "" {
				queryParams = "?"
			} else {
				queryParams += "&"
			}
			queryParams += fmt.Sprintf("filter[%s]=%s", key, value.(string))
		}
	}

	// Adicionar ordenação
	if v, ok := d.GetOk("sort"); ok {
		sortParams := v.(map[string]interface{})
		field, fieldOk := sortParams["field"]
		direction, directionOk := sortParams["direction"]

		if fieldOk {
			if queryParams == "" {
				queryParams = "?"
			} else {
				queryParams += "&"
			}

			if directionOk && direction.(string) == "desc" {
				queryParams += fmt.Sprintf("sort=-%s", field.(string))
			} else {
				queryParams += fmt.Sprintf("sort=%s", field.(string))
			}
		}
	}

	// Adicionar limite
	limit := 100
	if v, ok := d.GetOk("limit"); ok {
		limit = v.(int)
	}

	if queryParams == "" {
		queryParams = "?"
	} else {
		queryParams += "&"
	}
	queryParams += fmt.Sprintf("limit=%d", limit)

	// Adicionar org_id se especificado
	if v, ok := d.GetOk("org_id"); ok {
		orgID := v.(string)
		if queryParams == "" {
			queryParams = "?"
		} else {
			queryParams += "&"
		}
		queryParams += fmt.Sprintf("orgId=%s", orgID)
	}

	// Fazer a requisição
	tflog.Debug(ctx, fmt.Sprintf("Buscando políticas de autenticação: %s%s", url, queryParams))
	resp, err := c.DoRequest(http.MethodGet, url+queryParams, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao buscar políticas de autenticação: %v", err))
	}

	var policies AuthPoliciesResponse
	if err := json.Unmarshal(resp, &policies); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Transformar resultados para o formato esperado pelo Terraform
	authPolicies := make([]map[string]interface{}, 0, len(policies.Results))
	for _, policy := range policies.Results {
		policyMap := map[string]interface{}{
			"id":                 policy.ID,
			"name":               policy.Name,
			"description":        policy.Description,
			"type":               policy.Type,
			"status":             policy.Status,
			"priority":           policy.Priority,
			"apply_to_all_users": policy.ApplyToAllUsers,
			"effective_from":     policy.EffectiveFrom,
			"effective_until":    policy.EffectiveUntil,
			"created":            policy.Created,
			"updated":            policy.Updated,
		}

		// Converter settings para JSON
		if policy.Settings != nil {
			settingsJSON, err := json.Marshal(policy.Settings)
			if err != nil {
				return diag.FromErr(fmt.Errorf("erro ao serializar configurações: %v", err))
			}
			policyMap["settings"] = string(settingsJSON)
		}

		authPolicies = append(authPolicies, policyMap)
	}

	// Definir valores no state
	if err := d.Set("auth_policies", authPolicies); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir auth_policies: %v", err))
	}

	if err := d.Set("total_count", policies.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir total_count: %v", err))
	}

	// Gerar um ID único baseado no timestamp
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
