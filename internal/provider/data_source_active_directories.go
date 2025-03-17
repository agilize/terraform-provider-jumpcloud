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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// ActiveDirectoriesResponse representa a resposta da API para listagem de integrações AD
type ActiveDirectoriesResponse struct {
	Results []ActiveDirectory `json:"results"`
	Total   int64             `json:"totalCount"`
}

func dataSourceActiveDirectories() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceActiveDirectoriesRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtro por nome da integração",
			},
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtro por domínio",
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"regular", "gcs",
				}, false),
				Description: "Filtro por tipo de integração (regular, gcs)",
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"active", "pending", "error", "inactive",
				}, false),
				Description: "Filtro por status da integração",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Filtro por status de ativação",
			},
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 1000),
				Description:  "Número máximo de registros a retornar (1-1000)",
			},
			"skip": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Número de registros a pular (para paginação)",
			},
			"sort": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "name",
				ValidateFunc: validation.StringInSlice([]string{
					"name", "domain", "type", "status", "created", "updated",
				}, false),
				Description: "Campo para ordenação",
			},
			"sort_dir": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "asc",
				ValidateFunc: validation.StringInSlice([]string{
					"asc", "desc",
				}, false),
				Description: "Direção de ordenação (asc, desc)",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"directories": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID único da integração AD",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome da integração AD",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Descrição da integração AD",
						},
						"domain": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Domínio do Active Directory",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo de integração AD",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status atual da integração",
						},
						"use_ou": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indica se usa Unidade Organizacional específica",
						},
						"ou_path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Caminho da Unidade Organizacional",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indica se a integração está ativa",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da organização",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data de criação",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data da última atualização",
						},
					},
				},
				Description: "Lista de integrações AD encontradas",
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de integrações AD disponíveis",
			},
		},
	}
}

func dataSourceActiveDirectoriesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir parâmetros de consulta
	queryParams := constructActiveDirectoriesQueryParams(d)

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/activedirectories%s", queryParams)

	// Fazer requisição para buscar integrações AD
	tflog.Debug(ctx, "Buscando integrações de Active Directory")
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao buscar integrações AD: %v", err))
	}

	// Deserializar resposta
	var adResponse ActiveDirectoriesResponse
	if err := json.Unmarshal(resp, &adResponse); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Converter resultado para formato adequado para terraform
	directories := flattenActiveDirectories(adResponse.Results)

	// Definir valores no state
	if err := d.Set("directories", directories); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir directories: %v", err))
	}

	if err := d.Set("total", adResponse.Total); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir total: %v", err))
	}

	// Gerar ID exclusivo para este data source
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

// Constrói a string de parâmetros de consulta para a API
func constructActiveDirectoriesQueryParams(d *schema.ResourceData) string {
	params := "?"

	// Adicionar filtros
	if v, ok := d.GetOk("name"); ok {
		params += fmt.Sprintf("name=%s&", v.(string))
	}

	if v, ok := d.GetOk("domain"); ok {
		params += fmt.Sprintf("domain=%s&", v.(string))
	}

	if v, ok := d.GetOk("type"); ok {
		params += fmt.Sprintf("type=%s&", v.(string))
	}

	if v, ok := d.GetOk("status"); ok {
		params += fmt.Sprintf("status=%s&", v.(string))
	}

	if v, ok := d.GetOk("enabled"); ok {
		params += fmt.Sprintf("enabled=%t&", v.(bool))
	}

	// Adicionar paginação e ordenação
	params += fmt.Sprintf("limit=%d&", d.Get("limit").(int))
	params += fmt.Sprintf("skip=%d&", d.Get("skip").(int))
	params += fmt.Sprintf("sort=%s&", d.Get("sort").(string))
	params += fmt.Sprintf("sort_dir=%s&", d.Get("sort_dir").(string))

	// Adicionar orgId se fornecido
	if v, ok := d.GetOk("org_id"); ok {
		params += fmt.Sprintf("orgId=%s&", v.(string))
	}

	// Remover o último '&' se existir
	if len(params) > 1 && params[len(params)-1] == '&' {
		params = params[:len(params)-1]
	}

	return params
}

// Converte a resposta da API para o formato esperado pelo terraform
func flattenActiveDirectories(directories []ActiveDirectory) []interface{} {
	if directories == nil {
		return []interface{}{}
	}

	result := make([]interface{}, len(directories))
	for i, ad := range directories {
		dirMap := map[string]interface{}{
			"id":          ad.ID,
			"name":        ad.Name,
			"description": ad.Description,
			"domain":      ad.Domain,
			"type":        ad.Type,
			"status":      ad.Status,
			"use_ou":      ad.UseOU,
			"ou_path":     ad.OUPath,
			"enabled":     ad.Enabled,
			"org_id":      ad.OrgID,
			"created":     ad.Created,
			"updated":     ad.Updated,
		}
		result[i] = dirMap
	}

	return result
}
