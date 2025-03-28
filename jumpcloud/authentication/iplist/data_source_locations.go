package iplist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// IPLocationInfo representa informações de geolocalização de um IP
type IPLocationInfo struct {
	IP                string  `json:"ip"`
	CountryCode       string  `json:"countryCode"`
	CountryName       string  `json:"countryName"`
	RegionCode        string  `json:"regionCode"`
	RegionName        string  `json:"regionName"`
	City              string  `json:"city"`
	Latitude          float64 `json:"latitude"`
	Longitude         float64 `json:"longitude"`
	MetroCode         int     `json:"metroCode"`
	AreaCode          int     `json:"areaCode"`
	TimeZone          string  `json:"timeZone"`
	ContinentCode     string  `json:"continentCode"`
	PostalCode        string  `json:"postalCode"`
	ISP               string  `json:"isp"`
	Domain            string  `json:"domain"`
	AS                string  `json:"as"`
	ASName            string  `json:"asName"`
	Proxy             bool    `json:"proxy"`
	Mobile            bool    `json:"mobile"`
	ThreatLevel       string  `json:"threatLevel"`
	ThreatTypes       string  `json:"threatTypes"`
	ThreatClassifiers string  `json:"threatClassifiers"`
}

// IPLocationsResponse representa a resposta da API para consulta de geolocalização de IPs
type IPLocationsResponse struct {
	Results []IPLocationInfo `json:"results"`
}

func DataSourceLocations() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceLocationsRead,
		Schema: map[string]*schema.Schema{
			"ip_addresses": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Lista de endereços IP para consulta de geolocalização",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"locations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"country_code": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"country_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region_code": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"city": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"latitude": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"longitude": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"metro_code": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"area_code": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"time_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"continent_code": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"postal_code": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"isp": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"as": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"as_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"proxy": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"mobile": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"threat_level": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"threat_types": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"threat_classifiers": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Description: "Informações de geolocalização para cada endereço IP fornecido",
			},
		},
	}
}

func dataSourceLocationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obter a lista de IPs
	ipAddresses := d.Get("ip_addresses").([]interface{})
	if len(ipAddresses) == 0 {
		return diag.FromErr(fmt.Errorf("pelo menos um endereço IP deve ser fornecido"))
	}

	// Construir o payload para a requisição
	ips := make([]string, len(ipAddresses))
	for i, ip := range ipAddresses {
		ips[i] = ip.(string)
	}

	payload := map[string]interface{}{
		"ips": ips,
	}

	// Adicionar org_id se fornecido
	if v, ok := d.GetOk("org_id"); ok {
		payload["orgId"] = v.(string)
	}

	// Serializar para JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar payload: %v", err))
	}

	// Consultar geolocalização dos IPs via API
	tflog.Debug(ctx, "Consultando geolocalização dos IPs")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/ip-locations", payloadJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar geolocalização dos IPs: %v", err))
	}

	// Deserializar resposta
	var locationsResponse IPLocationsResponse
	if err := json.Unmarshal(resp, &locationsResponse); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Ordenar os resultados pelo endereço IP para consistência
	sort.Slice(locationsResponse.Results, func(i, j int) bool {
		return locationsResponse.Results[i].IP < locationsResponse.Results[j].IP
	})

	// Definir valores no state
	if err := d.Set("locations", flattenIPLocations(locationsResponse.Results)); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir locations: %v", err))
	}

	// Gerar ID para o data source (usando hash dos IPs consultados)
	d.SetId(fmt.Sprintf("%d-%s", len(ipAddresses), strings.Join(ips, "-")))

	return diags
}

// flattenIPLocations converte uma lista de IPLocationInfo para o formato do schema
func flattenIPLocations(locations []IPLocationInfo) []interface{} {
	if len(locations) == 0 {
		return make([]interface{}, 0)
	}

	result := make([]interface{}, len(locations))
	for i, loc := range locations {
		l := map[string]interface{}{
			"ip":                 loc.IP,
			"country_code":       loc.CountryCode,
			"country_name":       loc.CountryName,
			"region_code":        loc.RegionCode,
			"region_name":        loc.RegionName,
			"city":               loc.City,
			"latitude":           loc.Latitude,
			"longitude":          loc.Longitude,
			"metro_code":         loc.MetroCode,
			"area_code":          loc.AreaCode,
			"time_zone":          loc.TimeZone,
			"continent_code":     loc.ContinentCode,
			"postal_code":        loc.PostalCode,
			"isp":                loc.ISP,
			"domain":             loc.Domain,
			"as":                 loc.AS,
			"as_name":            loc.ASName,
			"proxy":              loc.Proxy,
			"mobile":             loc.Mobile,
			"threat_level":       loc.ThreatLevel,
			"threat_types":       loc.ThreatTypes,
			"threat_classifiers": loc.ThreatClassifiers,
		}
		result[i] = l
	}

	return result
}
