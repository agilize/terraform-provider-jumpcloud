package alerts

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

// AlertTemplate representa um template de alerta no JumpCloud
type AlertTemplate struct {
	ID                string                 `json:"_id,omitempty"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description,omitempty"`
	Type              string                 `json:"type"`
	DefaultSeverity   string                 `json:"defaultSeverity,omitempty"`
	DefaultConditions map[string]interface{} `json:"defaultConditions,omitempty"`
	Category          string                 `json:"category,omitempty"` // security, compliance, operations, etc.
	Scope             string                 `json:"scope,omitempty"`    // global, org, system, user, etc.
	PreConfigured     bool                   `json:"preConfigured"`      // template pré-configurado pelo JumpCloud
	OrgID             string                 `json:"orgId,omitempty"`
	Created           string                 `json:"created,omitempty"`
	Updated           string                 `json:"updated,omitempty"`
}

// AlertTemplatesResponse representa a resposta da API para listagem de templates de alertas
type AlertTemplatesResponse struct {
	Results     []AlertTemplate `json:"results"`
	TotalCount  int             `json:"totalCount"`
	NextPageURL string          `json:"nextPageUrl,omitempty"`
}

func DataSourceAlertTemplates() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAlertTemplatesRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar por tipo de alerta (system_metric, login_attempt, etc.)",
						},
						"category": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar por categoria (security, compliance, operations, etc.)",
						},
						"scope": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar por escopo (global, org, system, user, etc.)",
						},
						"search": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar por termo de busca (nome, descrição)",
						},
						"pre_configured": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Filtrar apenas templates pré-configurados pelo JumpCloud",
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
							ValidateFunc: validation.StringInSlice([]string{"name", "type", "category", "created", "updated"}, false),
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
				Description: "Número máximo de templates a serem retornados",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambiente multi-tenant",
			},
			"templates": {
				Type:     schema.TypeList,
				Computed: true,
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
							Description: "Tipo de alerta do template",
						},
						"default_severity": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Severidade padrão do template",
						},
						"default_conditions": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Condições padrão do template em formato JSON",
						},
						"category": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Categoria do template",
						},
						"scope": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Escopo do template",
						},
						"pre_configured": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Se o template é pré-configurado pelo JumpCloud",
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
				Description: "Número total de templates que correspondem aos filtros",
			},
		},
	}
}

func dataSourceAlertTemplatesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir URL base para a requisição
	url := "/api/v2/alert-templates"
	queryParams := ""

	// Aplicar filtros
	if filterList, ok := d.GetOk("filter"); ok && len(filterList.([]interface{})) > 0 {
		filter := filterList.([]interface{})[0].(map[string]interface{})

		if alertType, ok := filter["type"]; ok && alertType.(string) != "" {
			queryParams += fmt.Sprintf("&type=%s", alertType.(string))
		}

		if category, ok := filter["category"]; ok && category.(string) != "" {
			queryParams += fmt.Sprintf("&category=%s", category.(string))
		}

		if scope, ok := filter["scope"]; ok && scope.(string) != "" {
			queryParams += fmt.Sprintf("&scope=%s", scope.(string))
		}

		if search, ok := filter["search"]; ok && search.(string) != "" {
			queryParams += fmt.Sprintf("&search=%s", search.(string))
		}

		if preConfigured, ok := filter["pre_configured"]; ok {
			queryParams += fmt.Sprintf("&preConfigured=%t", preConfigured.(bool))
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
	tflog.Debug(ctx, fmt.Sprintf("Consultando templates de alertas: %s%s", url, queryParams))
	resp, err := c.DoRequest(http.MethodGet, url+queryParams, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar templates de alertas: %v", err))
	}

	// Deserializar resposta
	var response AlertTemplatesResponse
	if err := json.Unmarshal(resp, &response); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Formatar templates para o schema do Terraform
	templates := make([]interface{}, 0, len(response.Results))
	for _, template := range response.Results {
		var defaultConditionsJSON string
		if template.DefaultConditions != nil {
			condBytes, err := json.Marshal(template.DefaultConditions)
			if err != nil {
				return diag.FromErr(fmt.Errorf("erro ao serializar condições padrão: %v", err))
			}
			defaultConditionsJSON = string(condBytes)
		}

		templateMap := map[string]interface{}{
			"id":                 template.ID,
			"name":               template.Name,
			"description":        template.Description,
			"type":               template.Type,
			"default_severity":   template.DefaultSeverity,
			"default_conditions": defaultConditionsJSON,
			"category":           template.Category,
			"scope":              template.Scope,
			"pre_configured":     template.PreConfigured,
			"created":            template.Created,
			"updated":            template.Updated,
		}

		templates = append(templates, templateMap)
	}

	if err := d.Set("templates", templates); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir templates: %v", err))
	}

	if err := d.Set("total_count", response.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir total_count: %v", err))
	}

	// Definir ID do data source (usando timestamp para garantir unicidade)
	d.SetId(fmt.Sprintf("alert_templates_%d", time.Now().Unix()))

	return diags
}
