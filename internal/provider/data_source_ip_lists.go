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

// IPListsResponse representa a resposta da API para listagem de listas de IPs
type IPListsResponse struct {
	Results     []IPList `json:"results"`
	TotalCount  int      `json:"totalCount"`
	NextPageURL string   `json:"nextPageUrl,omitempty"`
}

func dataSourceIPLists() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIPListsRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Filtros para a listagem de listas de IPs",
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
				Description: "Número máximo de listas de IPs a serem retornadas",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"ip_lists": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Lista de listas de IPs",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da lista de IPs",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome da lista de IPs",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Descrição da lista de IPs",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo da lista de IPs (allow ou deny)",
						},
						"ips": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Lista de endereços IP/CIDR",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"address": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Endereço IP ou CIDR",
									},
									"description": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Descrição da entrada IP",
									},
								},
							},
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data de criação da lista de IPs",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data da última atualização da lista de IPs",
						},
					},
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de listas de IPs encontradas",
			},
		},
	}
}

func dataSourceIPListsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir a URL com query params
	url := "/api/v2/ip-lists"
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
	tflog.Debug(ctx, fmt.Sprintf("Buscando listas de IPs: %s%s", url, queryParams))
	resp, err := c.DoRequest(http.MethodGet, url+queryParams, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao buscar listas de IPs: %v", err))
	}

	var ipLists IPListsResponse
	if err := json.Unmarshal(resp, &ipLists); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Transformar resultados para o formato esperado pelo Terraform
	ipListsList := make([]map[string]interface{}, 0, len(ipLists.Results))
	for _, ipList := range ipLists.Results {
		ipListMap := map[string]interface{}{
			"id":          ipList.ID,
			"name":        ipList.Name,
			"description": ipList.Description,
			"type":        ipList.Type,
			"created":     ipList.Created,
			"updated":     ipList.Updated,
		}

		// Processar entradas de IPs
		if ipList.IPs != nil {
			ips := make([]map[string]interface{}, len(ipList.IPs))
			for i, ip := range ipList.IPs {
				ips[i] = map[string]interface{}{
					"address":     ip.Address,
					"description": ip.Description,
				}
			}
			ipListMap["ips"] = ips
		}

		ipListsList = append(ipListsList, ipListMap)
	}

	// Definir valores no state
	if err := d.Set("ip_lists", ipListsList); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir ip_lists: %v", err))
	}

	if err := d.Set("total_count", ipLists.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir total_count: %v", err))
	}

	// Gerar um ID único baseado no timestamp
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
