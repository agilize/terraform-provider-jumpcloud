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

// MDMPoliciesResponse representa a resposta da API para a consulta de políticas MDM
type MDMPoliciesResponse struct {
	Results     []MDMPolicy `json:"results"`
	TotalCount  int         `json:"totalCount"`
	NextPageURL string      `json:"nextPageUrl,omitempty"`
}

func dataSourceMDMPolicies() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMDMPoliciesRead,
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
							ValidateFunc: validation.StringInSlice([]string{"ios", "android", "windows", "all"}, false),
							Description:  "Filtrar por tipo de política (ios, android, windows, all)",
							Default:      "all",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Filtrar por estado de ativação (habilitada/desabilitada)",
						},
						"search": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Termo de busca para filtrar políticas (nome, descrição)",
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
							ValidateFunc: validation.StringInSlice([]string{"name", "type", "created", "updated"}, false),
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
				Description: "Número máximo de políticas a retornar",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambiente multi-tenant",
			},
			"policies": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da política MDM",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome da política MDM",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Descrição da política MDM",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo da política (ios, android, windows)",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Se a política está habilitada",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da organização da política",
						},
						"target_groups": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "IDs dos grupos alvo para aplicação da política",
						},
						"target_devices": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "IDs dos dispositivos alvo para aplicação da política",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data e hora de criação da política",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data e hora da última atualização da política",
						},
					},
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de políticas que correspondem aos filtros",
			},
		},
	}
}

func dataSourceMDMPoliciesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Construir URL base para a requisição
	url := "/api/v2/mdm/policies"
	queryParams := ""

	// Aplicar filtros
	if filterList, ok := d.GetOk("filter"); ok && len(filterList.([]interface{})) > 0 {
		filter := filterList.([]interface{})[0].(map[string]interface{})

		if policyType, ok := filter["type"]; ok && policyType.(string) != "all" {
			queryParams += fmt.Sprintf("&type=%s", policyType.(string))
		}

		if enabled, ok := filter["enabled"]; ok {
			queryParams += fmt.Sprintf("&enabled=%t", enabled.(bool))
		}

		if search, ok := filter["search"]; ok && search.(string) != "" {
			queryParams += fmt.Sprintf("&search=%s", search.(string))
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
	tflog.Debug(ctx, fmt.Sprintf("Consultando políticas MDM: %s%s", url, queryParams))
	resp, err := c.DoRequest(http.MethodGet, url+queryParams, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar políticas MDM: %v", err))
	}

	// Deserializar resposta
	var response MDMPoliciesResponse
	if err := json.Unmarshal(resp, &response); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Converter lista de políticas para o formato do Terraform
	policies := make([]map[string]interface{}, 0, len(response.Results))
	for _, policy := range response.Results {
		policyMap := map[string]interface{}{
			"id":          policy.ID,
			"name":        policy.Name,
			"description": policy.Description,
			"type":        policy.Type,
			"enabled":     policy.Enabled,
			"created":     policy.Created,
			"updated":     policy.Updated,
		}

		// Campos opcionais
		if policy.OrgID != "" {
			policyMap["org_id"] = policy.OrgID
		}
		if len(policy.TargetGroups) > 0 {
			policyMap["target_groups"] = policy.TargetGroups
		}
		if len(policy.TargetDevices) > 0 {
			policyMap["target_devices"] = policy.TargetDevices
		}

		policies = append(policies, policyMap)
	}

	// Definir valores no state
	if err := d.Set("policies", policies); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir políticas no state: %v", err))
	}

	if err := d.Set("total_count", response.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir total_count no state: %v", err))
	}

	// Definir ID único para o data source (baseado no timestamp atual)
	d.SetId(fmt.Sprintf("mdm-policies-%d", time.Now().Unix()))

	return diags
}
