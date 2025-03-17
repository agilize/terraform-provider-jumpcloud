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

// AppCatalogCategoriesResponse representa a resposta da API para a consulta de categorias do catálogo
type AppCatalogCategoriesResponse struct {
	Results     []AppCatalogCategory `json:"results"`
	TotalCount  int                  `json:"totalCount"`
	NextPageURL string               `json:"nextPageUrl,omitempty"`
}

func dataSourceAppCatalogCategories() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAppCatalogCategoriesRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"parent_category": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar por categoria pai (ID)",
						},
						"search": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Termo de busca para filtrar categorias (nome, descrição)",
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
							ValidateFunc: validation.StringInSlice([]string{"name", "displayOrder", "created", "updated"}, false),
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
				Description: "Número máximo de categorias a retornar",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambiente multi-tenant",
			},
			"categories": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da categoria",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome da categoria",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Descrição da categoria",
						},
						"display_order": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Ordem de exibição da categoria",
						},
						"parent_category": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da categoria pai (para subcategorias)",
						},
						"icon_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL do ícone da categoria",
						},
						"applications": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "IDs das aplicações que pertencem a esta categoria",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data de criação da categoria",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data da última atualização da categoria",
						},
					},
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de categorias que correspondem aos filtros",
			},
		},
	}
}

func dataSourceAppCatalogCategoriesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Construir URL base para a requisição
	url := "/api/v2/appcatalog/categories"
	queryParams := ""

	// Aplicar filtros
	if filterList, ok := d.GetOk("filter"); ok && len(filterList.([]interface{})) > 0 {
		filter := filterList.([]interface{})[0].(map[string]interface{})

		if parentCategory, ok := filter["parent_category"]; ok && parentCategory.(string) != "" {
			queryParams += fmt.Sprintf("&parentCategory=%s", parentCategory.(string))
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
	tflog.Debug(ctx, fmt.Sprintf("Consultando categorias do catálogo: %s%s", url, queryParams))
	resp, err := c.DoRequest(http.MethodGet, url+queryParams, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar categorias do catálogo: %v", err))
	}

	// Deserializar resposta
	var response AppCatalogCategoriesResponse
	if err := json.Unmarshal(resp, &response); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Converter lista de categorias para o formato do Terraform
	categories := make([]map[string]interface{}, 0, len(response.Results))
	for _, category := range response.Results {
		categoryMap := map[string]interface{}{
			"id":              category.ID,
			"name":            category.Name,
			"description":     category.Description,
			"display_order":   category.DisplayOrder,
			"parent_category": category.ParentCategory,
			"icon_url":        category.IconURL,
			"applications":    category.Applications,
			"created":         category.Created,
			"updated":         category.Updated,
		}

		categories = append(categories, categoryMap)
	}

	// Definir valores no state
	if err := d.Set("categories", categories); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir categorias no state: %v", err))
	}

	if err := d.Set("total_count", response.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir total_count no state: %v", err))
	}

	// Definir ID único para o data source (baseado no timestamp atual)
	d.SetId(fmt.Sprintf("app-catalog-categories-%d", time.Now().Unix()))

	return diags
}
