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

// AuthPolicyTemplate representa um template de política de autenticação no JumpCloud
type AuthPolicyTemplate struct {
	ID          string                 `json:"_id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type"` // mfa, password, lockout, session
	Settings    map[string]interface{} `json:"settings,omitempty"`
	Category    string                 `json:"category,omitempty"` // security, compliance, custom, etc.
	OrgID       string                 `json:"orgId,omitempty"`
	Created     string                 `json:"created,omitempty"`
	Updated     string                 `json:"updated,omitempty"`
}

// AuthPolicyTemplatesResponse representa a resposta da API para listagem de templates de políticas
type AuthPolicyTemplatesResponse struct {
	Results     []AuthPolicyTemplate `json:"results"`
	TotalCount  int                  `json:"totalCount"`
	NextPageURL string               `json:"nextPageUrl,omitempty"`
}

func dataSourceAuthPolicyTemplates() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAuthPolicyTemplatesRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Filtros para a listagem de templates",
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
				Description: "Número máximo de templates a serem retornados",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"templates": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Lista de templates de políticas de autenticação",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do template",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome do template",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Descrição do template",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo do template (mfa, password, lockout, session)",
						},
						"settings": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Configurações do template em formato JSON",
						},
						"category": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Categoria do template (security, compliance, custom, etc.)",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data de criação do template",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data da última atualização do template",
						},
					},
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de templates encontrados",
			},
		},
	}
}

func dataSourceAuthPolicyTemplatesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Construir a URL com query params
	url := "/api/v2/auth-policy-templates"
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
	tflog.Debug(ctx, fmt.Sprintf("Buscando templates de políticas: %s%s", url, queryParams))
	resp, err := c.DoRequest(http.MethodGet, url+queryParams, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao buscar templates de políticas: %v", err))
	}

	var templates AuthPolicyTemplatesResponse
	if err := json.Unmarshal(resp, &templates); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Transformar resultados para o formato esperado pelo Terraform
	templatesList := make([]map[string]interface{}, 0, len(templates.Results))
	for _, template := range templates.Results {
		templateMap := map[string]interface{}{
			"id":          template.ID,
			"name":        template.Name,
			"description": template.Description,
			"type":        template.Type,
			"category":    template.Category,
			"created":     template.Created,
			"updated":     template.Updated,
		}

		// Converter settings para JSON
		if template.Settings != nil {
			settingsJSON, err := json.Marshal(template.Settings)
			if err != nil {
				return diag.FromErr(fmt.Errorf("erro ao serializar configurações: %v", err))
			}
			templateMap["settings"] = string(settingsJSON)
		}

		templatesList = append(templatesList, templateMap)
	}

	// Definir valores no state
	if err := d.Set("templates", templatesList); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir templates: %v", err))
	}

	if err := d.Set("total_count", templates.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir total_count: %v", err))
	}

	// Gerar um ID único baseado no timestamp
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
