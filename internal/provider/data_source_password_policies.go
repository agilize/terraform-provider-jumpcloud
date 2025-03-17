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

// PasswordPolicyItem representa uma política de senha do JumpCloud no data source
type PasswordPolicyItem struct {
	ID                        string   `json:"_id"`
	Name                      string   `json:"name"`
	Description               string   `json:"description,omitempty"`
	Status                    string   `json:"status"`
	MinLength                 int      `json:"minLength"`
	MaxLength                 int      `json:"maxLength,omitempty"`
	RequireUppercase          bool     `json:"requireUppercase"`
	RequireLowercase          bool     `json:"requireLowercase"`
	RequireNumber             bool     `json:"requireNumber"`
	RequireSymbol             bool     `json:"requireSymbol"`
	MinimumAge                int      `json:"minimumAge,omitempty"`
	ExpirationTime            int      `json:"expirationTime,omitempty"`
	ExpirationWarningTime     int      `json:"expirationWarningTime,omitempty"`
	DisallowPreviousPasswords int      `json:"disallowPreviousPasswords,omitempty"`
	DisallowCommonPasswords   bool     `json:"disallowCommonPasswords"`
	DisallowUsername          bool     `json:"disallowUsername"`
	DisallowNameAndEmail      bool     `json:"disallowNameAndEmail"`
	DisallowPasswordsFromList bool     `json:"disallowPasswordsFromList"`
	Scope                     string   `json:"scope,omitempty"`
	TargetResources           []string `json:"targetResources,omitempty"`
	OrgID                     string   `json:"orgId,omitempty"`
	Created                   string   `json:"created"`
	Updated                   string   `json:"updated"`
}

// PasswordPoliciesResponse representa a resposta da API para listagem de políticas de senha
type PasswordPoliciesResponse struct {
	Results    []PasswordPolicyItem `json:"results"`
	TotalCount int                  `json:"totalCount"`
}

func dataSourcePasswordPolicies() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePasswordPoliciesRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar por nome da política",
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar por status (active, inactive)",
			},
			"scope": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar por escopo da política (organization, directory, ou specific)",
			},
			"search": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar políticas por texto em nome ou descrição",
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     100,
				Description: "Número máximo de políticas a serem retornadas",
			},
			"skip": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Número de políticas a serem ignoradas",
			},
			"sort": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "name",
				Description: "Campo para ordenação dos resultados",
			},
			"sort_dir": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "asc",
				Description: "Direção da ordenação (asc ou desc)",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"policies": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Lista de políticas de senha encontradas",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da política de senha",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome da política de senha",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Descrição da política de senha",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status da política (active, inactive)",
						},
						"min_length": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Comprimento mínimo da senha",
						},
						"max_length": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Comprimento máximo da senha",
						},
						"require_uppercase": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Exige letra maiúscula",
						},
						"require_lowercase": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Exige letra minúscula",
						},
						"require_number": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Exige número",
						},
						"require_symbol": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Exige símbolo",
						},
						"minimum_age": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Idade mínima da senha em dias",
						},
						"expiration_time": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Tempo de expiração da senha em dias",
						},
						"expiration_warning_time": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Tempo de aviso antes da expiração em dias",
						},
						"disallow_previous_passwords": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Número de senhas anteriores não permitidas",
						},
						"disallow_common_passwords": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Proíbe senhas comuns",
						},
						"disallow_username": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Proíbe o uso do nome de usuário na senha",
						},
						"disallow_name_and_email": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Proíbe o uso do nome ou email na senha",
						},
						"disallow_passwords_from_list": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Proíbe senhas de uma lista personalizada",
						},
						"scope": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Escopo da política (organization, directory, specific)",
						},
						"target_resources": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "IDs dos recursos alvo (quando scope é specific)",
							Elem:        &schema.Schema{Type: schema.TypeString},
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
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de políticas encontradas",
			},
		},
	}
}

func dataSourcePasswordPoliciesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir parâmetros de consulta
	queryParams := constructPasswordPoliciesQueryParams(d)

	// Construir URL com parâmetros
	url := fmt.Sprintf("/api/v2/password-policies?%s", queryParams)

	// Buscar políticas via API
	tflog.Debug(ctx, fmt.Sprintf("Listando políticas de senha com parâmetros: %s", queryParams))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao listar políticas de senha: %v", err))
	}

	// Deserializar resposta
	var policiesResp PasswordPoliciesResponse
	if err := json.Unmarshal(resp, &policiesResp); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Converter políticas para formato Terraform
	tfPolicies := flattenPasswordPolicies(policiesResp.Results)
	if err := d.Set("policies", tfPolicies); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir policies: %v", err))
	}

	d.Set("total", policiesResp.TotalCount)

	// Gerar ID único para o data source
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

// Função auxiliar para construir os parâmetros de consulta
func constructPasswordPoliciesQueryParams(d *schema.ResourceData) string {
	params := ""

	// Adicionar filtros
	if v, ok := d.GetOk("name"); ok {
		params += fmt.Sprintf("name=%s&", v.(string))
	}

	if v, ok := d.GetOk("status"); ok {
		params += fmt.Sprintf("status=%s&", v.(string))
	}

	if v, ok := d.GetOk("scope"); ok {
		params += fmt.Sprintf("scope=%s&", v.(string))
	}

	if v, ok := d.GetOk("search"); ok {
		params += fmt.Sprintf("search=%s&", v.(string))
	}

	// Adicionar parâmetros de paginação e ordenação
	params += fmt.Sprintf("limit=%d&", d.Get("limit").(int))
	params += fmt.Sprintf("skip=%d&", d.Get("skip").(int))
	params += fmt.Sprintf("sort=%s&", d.Get("sort").(string))
	params += fmt.Sprintf("sort_dir=%s&", d.Get("sort_dir").(string))

	// Adicionar org_id se fornecido
	if v, ok := d.GetOk("org_id"); ok {
		params += fmt.Sprintf("org_id=%s&", v.(string))
	}

	// Remover último & se existir
	if len(params) > 0 && params[len(params)-1] == '&' {
		params = params[:len(params)-1]
	}

	return params
}

// Função auxiliar para converter políticas para formato adequado ao Terraform
func flattenPasswordPolicies(policies []PasswordPolicyItem) []map[string]interface{} {
	result := make([]map[string]interface{}, len(policies))

	for i, policy := range policies {
		policyMap := map[string]interface{}{
			"id":                           policy.ID,
			"name":                         policy.Name,
			"description":                  policy.Description,
			"status":                       policy.Status,
			"min_length":                   policy.MinLength,
			"require_uppercase":            policy.RequireUppercase,
			"require_lowercase":            policy.RequireLowercase,
			"require_number":               policy.RequireNumber,
			"require_symbol":               policy.RequireSymbol,
			"disallow_common_passwords":    policy.DisallowCommonPasswords,
			"disallow_username":            policy.DisallowUsername,
			"disallow_name_and_email":      policy.DisallowNameAndEmail,
			"disallow_passwords_from_list": policy.DisallowPasswordsFromList,
			"created":                      policy.Created,
			"updated":                      policy.Updated,
		}

		// Campos opcionais
		if policy.MaxLength > 0 {
			policyMap["max_length"] = policy.MaxLength
		}

		if policy.MinimumAge > 0 {
			policyMap["minimum_age"] = policy.MinimumAge
		}

		if policy.ExpirationTime > 0 {
			policyMap["expiration_time"] = policy.ExpirationTime
		}

		if policy.ExpirationWarningTime > 0 {
			policyMap["expiration_warning_time"] = policy.ExpirationWarningTime
		}

		if policy.DisallowPreviousPasswords > 0 {
			policyMap["disallow_previous_passwords"] = policy.DisallowPreviousPasswords
		}

		if policy.Scope != "" {
			policyMap["scope"] = policy.Scope
		}

		if policy.OrgID != "" {
			policyMap["org_id"] = policy.OrgID
		}

		if policy.TargetResources != nil && len(policy.TargetResources) > 0 {
			policyMap["target_resources"] = policy.TargetResources
		}

		result[i] = policyMap
	}

	return result
}
