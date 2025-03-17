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

// AppCatalogApplicationsResponse representa a resposta da API para a consulta de aplicativos do catálogo
type AppCatalogApplicationsResponse struct {
	Results     []AppCatalogApplication `json:"results"`
	TotalCount  int                     `json:"totalCount"`
	NextPageURL string                  `json:"nextPageUrl,omitempty"`
}

func dataSourceAppCatalogApplications() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAppCatalogApplicationsRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"app_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"web", "mobile", "desktop", "all"}, false),
							Description:  "Filtrar por tipo de aplicação (web, mobile, desktop, all)",
							Default:      "all",
						},
						"status": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"active", "inactive", "draft", "all"}, false),
							Description:  "Filtrar por status da aplicação (active, inactive, draft, all)",
							Default:      "all",
						},
						"visibility": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"public", "private", "all"}, false),
							Description:  "Filtrar por visibilidade da aplicação (public, private, all)",
							Default:      "all",
						},
						"categories": {
							Type:        schema.TypeList,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Lista de IDs de categorias para filtrar",
						},
						"platform_support": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"ios", "android", "windows", "macos", "web", "all",
							}, false),
							Description: "Filtrar por plataforma suportada",
							Default:     "all",
						},
						"search": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Termo de busca para filtrar aplicações (nome, descrição)",
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
							ValidateFunc: validation.StringInSlice([]string{"name", "publisher", "appType", "created", "updated"}, false),
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
				Description: "Número máximo de aplicações a retornar",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambiente multi-tenant",
			},
			"applications": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da aplicação",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome da aplicação",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Descrição da aplicação",
						},
						"icon_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL do ícone da aplicação",
						},
						"app_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo da aplicação (web, mobile, desktop)",
						},
						"publisher": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Editor/Publicador da aplicação",
						},
						"version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Versão da aplicação",
						},
						"license": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo de licença da aplicação",
						},
						"install_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo de instalação da aplicação",
						},
						"app_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL da aplicação (para web apps)",
						},
						"app_store_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL da aplicação na loja de aplicativos",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status da aplicação",
						},
						"visibility": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Visibilidade da aplicação",
						},
						"categories": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Categorias da aplicação",
						},
						"platform_support": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Plataformas suportadas pela aplicação",
						},
						"tags": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Tags associadas à aplicação",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data de criação da aplicação",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data da última atualização da aplicação",
						},
					},
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de aplicações que correspondem aos filtros",
			},
		},
	}
}

func dataSourceAppCatalogApplicationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Construir URL base para a requisição
	url := "/api/v2/appcatalog/applications"
	queryParams := ""

	// Aplicar filtros
	if filterList, ok := d.GetOk("filter"); ok && len(filterList.([]interface{})) > 0 {
		filter := filterList.([]interface{})[0].(map[string]interface{})

		if appType, ok := filter["app_type"]; ok && appType.(string) != "all" {
			queryParams += fmt.Sprintf("&appType=%s", appType.(string))
		}

		if status, ok := filter["status"]; ok && status.(string) != "all" {
			queryParams += fmt.Sprintf("&status=%s", status.(string))
		}

		if visibility, ok := filter["visibility"]; ok && visibility.(string) != "all" {
			queryParams += fmt.Sprintf("&visibility=%s", visibility.(string))
		}

		if platformSupport, ok := filter["platform_support"]; ok && platformSupport.(string) != "all" {
			queryParams += fmt.Sprintf("&platformSupport=%s", platformSupport.(string))
		}

		if search, ok := filter["search"]; ok && search.(string) != "" {
			queryParams += fmt.Sprintf("&search=%s", search.(string))
		}

		if categories, ok := filter["categories"]; ok {
			categoryList := categories.([]interface{})
			for _, cat := range categoryList {
				queryParams += fmt.Sprintf("&categories=%s", cat.(string))
			}
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
	tflog.Debug(ctx, fmt.Sprintf("Consultando aplicações do catálogo: %s%s", url, queryParams))
	resp, err := c.DoRequest(http.MethodGet, url+queryParams, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar aplicações do catálogo: %v", err))
	}

	// Deserializar resposta
	var response AppCatalogApplicationsResponse
	if err := json.Unmarshal(resp, &response); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Converter lista de aplicações para o formato do Terraform
	applications := make([]map[string]interface{}, 0, len(response.Results))
	for _, app := range response.Results {
		appMap := map[string]interface{}{
			"id":               app.ID,
			"name":             app.Name,
			"description":      app.Description,
			"icon_url":         app.IconURL,
			"app_type":         app.AppType,
			"publisher":        app.Publisher,
			"version":          app.Version,
			"license":          app.License,
			"install_type":     app.InstallType,
			"app_url":          app.AppURL,
			"app_store_url":    app.AppStoreURL,
			"status":           app.Status,
			"visibility":       app.Visibility,
			"categories":       app.Categories,
			"platform_support": app.PlatformSupport,
			"tags":             app.Tags,
			"created":          app.Created,
			"updated":          app.Updated,
		}

		applications = append(applications, appMap)
	}

	// Definir valores no state
	if err := d.Set("applications", applications); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir aplicações no state: %v", err))
	}

	if err := d.Set("total_count", response.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir total_count no state: %v", err))
	}

	// Definir ID único para o data source (baseado no timestamp atual)
	d.SetId(fmt.Sprintf("app-catalog-applications-%d", time.Now().Unix()))

	return diags
}
