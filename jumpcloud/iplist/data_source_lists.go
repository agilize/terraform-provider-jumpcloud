package iplist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

func DataSourceLists() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceListsRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar pelo nome da lista de IPs",
						},
						"type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar pelo tipo de lista de IPs (allow ou deny)",
						},
					},
				},
				Description: "Filtros para limitar os resultados",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"ip_lists": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ips": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"address": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"description": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"created": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"updated": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Description: "Lista de listas de IPs encontradas",
			},
		},
	}
}

func dataSourceListsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir URL para requisição
	url := "/api/v2/ip-lists"

	// Adicionar org_id se fornecido
	if v, ok := d.GetOk("org_id"); ok {
		url = fmt.Sprintf("%s?orgId=%s", url, v.(string))
	}

	// Buscar listas de IPs via API
	tflog.Debug(ctx, "Buscando listas de IPs")
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao buscar listas de IPs: %v", err))
	}

	// Deserializar resposta
	var ipListsResponse IPListsResponse
	if err := json.Unmarshal(resp, &ipListsResponse); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Filtrar resultados se houver filtros
	var filteredIPLists []IPList
	if filters, ok := d.GetOk("filter"); ok {
		filterList := filters.([]interface{})
		if len(filterList) > 0 {
			filter := filterList[0].(map[string]interface{})
			filteredIPLists = filterIPLists(ipListsResponse.Results, filter)
		} else {
			filteredIPLists = ipListsResponse.Results
		}
	} else {
		filteredIPLists = ipListsResponse.Results
	}

	// Definir valores no state
	if err := d.Set("ip_lists", flattenIPLists(filteredIPLists)); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir ip_lists: %v", err))
	}

	// Gerar ID para o data source (usando hash dos resultados)
	d.SetId(fmt.Sprintf("%d", len(filteredIPLists)))

	return diags
}

// filterIPLists filtra as listas de IPs pelos critérios fornecidos
func filterIPLists(ipLists []IPList, filter map[string]interface{}) []IPList {
	var filtered []IPList

	for _, ipList := range ipLists {
		match := true

		if v, ok := filter["name"]; ok && v.(string) != "" && ipList.Name != v.(string) {
			match = false
		}

		if v, ok := filter["type"]; ok && v.(string) != "" && ipList.Type != v.(string) {
			match = false
		}

		if match {
			filtered = append(filtered, ipList)
		}
	}

	return filtered
}

// flattenIPLists converte uma lista de IPList para o formato do schema
func flattenIPLists(ipLists []IPList) []interface{} {
	if len(ipLists) == 0 {
		return make([]interface{}, 0)
	}

	result := make([]interface{}, len(ipLists))
	for i, ipList := range ipLists {
		l := map[string]interface{}{
			"id":          ipList.ID,
			"name":        ipList.Name,
			"description": ipList.Description,
			"type":        ipList.Type,
			"created":     ipList.Created,
			"updated":     ipList.Updated,
			"ips":         flattenIPAddressEntriesList(ipList.IPs),
		}
		result[i] = l
	}

	return result
}

// flattenIPAddressEntriesList converte uma lista de IPAddressEntry para o formato do schema
func flattenIPAddressEntriesList(entries []IPAddressEntry) []interface{} {
	if entries == nil {
		return []interface{}{}
	}

	flattened := make([]interface{}, len(entries))
	for i, entry := range entries {
		entryMap := map[string]interface{}{
			"address":     entry.Address,
			"description": entry.Description,
		}
		flattened[i] = entryMap
	}

	return flattened
}
