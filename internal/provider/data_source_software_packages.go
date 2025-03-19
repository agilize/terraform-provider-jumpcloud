package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// SoftwarePackageListItem representa um item na lista de pacotes de software
type SoftwarePackageListItem struct {
	ID          string   `json:"_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Type        string   `json:"type"` // windows, macos, linux
	Status      string   `json:"status"`
	FileSize    int64    `json:"fileSize"`
	Tags        []string `json:"tags"`
	Created     string   `json:"created"`
	Updated     string   `json:"updated"`
}

// SoftwarePackagesResponse representa a resposta da API de listagem de pacotes
type SoftwarePackagesResponse struct {
	Results    []SoftwarePackageListItem `json:"results"`
	TotalCount int                       `json:"totalCount"`
}

func dataSourceSoftwarePackages() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSoftwarePackagesRead,
		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"windows", "macos", "linux",
				}, false),
				Description: "Filtrar por tipo de sistema operacional (windows, macos, linux)",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar por nome (pesquisa parcial)",
			},
			"version": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar por versão exata",
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"active", "inactive", "processing", "error",
				}, false),
				Description: "Filtrar por status",
			},
			"has_tag": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar por tag específica",
			},
			"sort": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "name",
				ValidateFunc: validation.StringInSlice([]string{
					"name", "type", "version", "created", "updated",
				}, false),
				Description: "Campo para ordenação dos resultados",
			},
			"sort_dir": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "asc",
				ValidateFunc: validation.StringInSlice([]string{
					"asc", "desc",
				}, false),
				Description: "Direção da ordenação (asc, desc)",
			},
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      50,
				ValidateFunc: validation.IntBetween(1, 1000),
				Description:  "Número máximo de resultados (1-1000)",
			},
			"skip": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Número de resultados a pular (para paginação)",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"packages": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Lista de pacotes de software",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do pacote",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome do pacote",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Descrição do pacote",
						},
						"version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Versão do pacote",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo de sistema operacional",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status do pacote",
						},
						"file_size": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Tamanho do arquivo em bytes",
						},
						"tags": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Tags associadas ao pacote",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data de criação do pacote",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data da última atualização do pacote",
						},
					},
				},
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de pacotes que correspondem aos filtros",
			},
		},
	}
}

func dataSourceSoftwarePackagesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir query parameters para filtros
	query := url.Values{}

	// Adicionar filtros
	if v, ok := d.GetOk("type"); ok {
		query.Add("type", v.(string))
	}
	if v, ok := d.GetOk("name"); ok {
		query.Add("name", v.(string))
	}
	if v, ok := d.GetOk("version"); ok {
		query.Add("version", v.(string))
	}
	if v, ok := d.GetOk("status"); ok {
		query.Add("status", v.(string))
	}
	if v, ok := d.GetOk("has_tag"); ok {
		query.Add("hasTag", v.(string))
	}

	// Adicionar parâmetros de paginação e ordenação
	query.Add("sort", d.Get("sort").(string))
	query.Add("sortDir", d.Get("sort_dir").(string))
	query.Add("limit", strconv.Itoa(d.Get("limit").(int)))
	query.Add("skip", strconv.Itoa(d.Get("skip").(int)))

	// Adicionar orgId se disponível
	if v, ok := d.GetOk("org_id"); ok {
		query.Add("orgId", v.(string))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/software/packages?%s", query.Encode())

	// Fazer requisição para listar pacotes
	tflog.Debug(ctx, "Listando pacotes de software")
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao listar pacotes de software: %v", err))
	}

	// Deserializar resposta
	var packagesResp SoftwarePackagesResponse
	if err := json.Unmarshal(resp, &packagesResp); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir ID do data source
	d.SetId(fmt.Sprintf("software-packages-%d", time.Now().Unix()))

	// Mapear valores para o schema
	d.Set("total", packagesResp.TotalCount)

	// Processar pacotes
	packages := make([]map[string]interface{}, len(packagesResp.Results))
	for i, pkg := range packagesResp.Results {
		packageMap := map[string]interface{}{
			"id":          pkg.ID,
			"name":        pkg.Name,
			"description": pkg.Description,
			"version":     pkg.Version,
			"type":        pkg.Type,
			"status":      pkg.Status,
			"file_size":   pkg.FileSize,
			"created":     pkg.Created,
			"updated":     pkg.Updated,
		}

		// Converter tags para formato adequado ao Terraform
		if pkg.Tags != nil {
			packageMap["tags"] = pkg.Tags
		} else {
			packageMap["tags"] = []string{}
		}

		packages[i] = packageMap
	}

	if err := d.Set("packages", packages); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir packages: %v", err))
	}

	return diags
}
