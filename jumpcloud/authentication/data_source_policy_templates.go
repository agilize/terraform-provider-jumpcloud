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

func DataSourcePolicyTemplates() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePolicyTemplatesRead,
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
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da organização",
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

func dataSourcePolicyTemplatesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Consultar templates de política via API
	tflog.Debug(ctx, "Consultando templates de políticas de autenticação")
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/auth-policy-templates%s", queryParams), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar templates de políticas: %v", err))
	}

	// Deserializar resposta
	var templatesResp AuthPolicyTemplatesResponse
	if err := json.Unmarshal(resp, &templatesResp); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	d.Set("total_count", templatesResp.TotalCount)

	// Preparar resultados
	templates := make([]map[string]interface{}, 0, len(templatesResp.Results))
	for _, template := range templatesResp.Results {
		// Serializar settings para JSON
		settingsJSON, err := json.Marshal(template.Settings)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao serializar settings: %v", err))
		}

		tmpl := map[string]interface{}{
			"id":          template.ID,
			"name":        template.Name,
			"description": template.Description,
			"type":        template.Type,
			"settings":    string(settingsJSON),
			"category":    template.Category,
			"created":     template.Created,
			"updated":     template.Updated,
		}

		if template.OrgID != "" {
			tmpl["org_id"] = template.OrgID
		}

		templates = append(templates, tmpl)
	}

	if err := d.Set("templates", templates); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir templates: %v", err))
	}

	// Definir ID do recurso de dados (timestamp para garantir unicidade)
	d.SetId(fmt.Sprintf("auth_policy_templates_%d", time.Now().Unix()))

	return diags
}
