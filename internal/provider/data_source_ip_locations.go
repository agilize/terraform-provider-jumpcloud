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

// IPLocationInfo representa informações de geolocalização de um endereço IP
type IPLocationInfo struct {
	IP           string  `json:"ip"`
	CountryCode  string  `json:"countryCode,omitempty"`
	CountryName  string  `json:"countryName,omitempty"`
	RegionCode   string  `json:"regionCode,omitempty"`
	RegionName   string  `json:"regionName,omitempty"`
	City         string  `json:"city,omitempty"`
	ZipCode      string  `json:"zipCode,omitempty"`
	Latitude     float64 `json:"latitude,omitempty"`
	Longitude    float64 `json:"longitude,omitempty"`
	ISP          string  `json:"isp,omitempty"`
	Organization string  `json:"organization,omitempty"`
	TimeZone     string  `json:"timeZone,omitempty"`
	Continent    string  `json:"continent,omitempty"`
	IsTor        bool    `json:"isTor,omitempty"`
	IsProxy      bool    `json:"isProxy,omitempty"`
	IsVPN        bool    `json:"isVpn,omitempty"`
	IsHosting    bool    `json:"isHosting,omitempty"`
	Threat       int     `json:"threat,omitempty"` // Nível de ameaça de 0 a 100
}

// IPLocationsResponse representa a resposta da API para a consulta de localizações de IPs
type IPLocationsResponse struct {
	Results []IPLocationInfo `json:"results"`
}

func dataSourceIPLocations() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIPLocationsRead,
		Schema: map[string]*schema.Schema{
			"ip_addresses": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Lista de endereços IP para consulta de geolocalização",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"locations": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Informações de geolocalização dos endereços IP",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Endereço IP consultado",
						},
						"country_code": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Código do país (ISO)",
						},
						"country_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome do país",
						},
						"region_code": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Código da região/estado",
						},
						"region_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome da região/estado",
						},
						"city": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome da cidade",
						},
						"zip_code": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Código postal",
						},
						"latitude": {
							Type:        schema.TypeFloat,
							Computed:    true,
							Description: "Latitude",
						},
						"longitude": {
							Type:        schema.TypeFloat,
							Computed:    true,
							Description: "Longitude",
						},
						"isp": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Provedor de serviços de internet",
						},
						"organization": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Organização proprietária do IP",
						},
						"time_zone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Fuso horário",
						},
						"continent": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Continente",
						},
						"is_tor": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Se o IP é um nó de saída Tor",
						},
						"is_proxy": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Se o IP é um proxy conhecido",
						},
						"is_vpn": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Se o IP pertence a um serviço VPN conhecido",
						},
						"is_hosting": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Se o IP pertence a um provedor de hospedagem",
						},
						"threat": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Nível de ameaça de 0 a 100",
						},
					},
				},
			},
		},
	}
}

func dataSourceIPLocationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obter a lista de IPs para consulta
	ipSet := d.Get("ip_addresses").(*schema.Set)
	ips := ipSet.List()
	if len(ips) == 0 {
		return diag.FromErr(fmt.Errorf("pelo menos um endereço IP deve ser fornecido"))
	}

	// Preparar a lista de IPs para consulta
	ipAddresses := make([]string, len(ips))
	for i, ip := range ips {
		ipAddresses[i] = ip.(string)
	}

	// Construir o corpo da requisição
	requestBody := map[string]interface{}{
		"ips": ipAddresses,
	}

	// Adicionar org_id se especificado
	if v, ok := d.GetOk("org_id"); ok {
		requestBody["orgId"] = v.(string)
	}

	// Serializar para JSON
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar requisição: %v", err))
	}

	// Fazer a requisição
	tflog.Debug(ctx, "Consultando localização de IPs")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/ip-locations", requestBodyJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar localização de IPs: %v", err))
	}

	var ipLocations IPLocationsResponse
	if err := json.Unmarshal(resp, &ipLocations); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Transformar resultados para o formato esperado pelo Terraform
	locations := make([]map[string]interface{}, 0, len(ipLocations.Results))
	for _, location := range ipLocations.Results {
		locationMap := map[string]interface{}{
			"ip":           location.IP,
			"country_code": location.CountryCode,
			"country_name": location.CountryName,
			"region_code":  location.RegionCode,
			"region_name":  location.RegionName,
			"city":         location.City,
			"zip_code":     location.ZipCode,
			"latitude":     location.Latitude,
			"longitude":    location.Longitude,
			"isp":          location.ISP,
			"organization": location.Organization,
			"time_zone":    location.TimeZone,
			"continent":    location.Continent,
			"is_tor":       location.IsTor,
			"is_proxy":     location.IsProxy,
			"is_vpn":       location.IsVPN,
			"is_hosting":   location.IsHosting,
			"threat":       location.Threat,
		}

		locations = append(locations, locationMap)
	}

	// Definir valores no state
	if err := d.Set("locations", locations); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir locations: %v", err))
	}

	// Gerar um ID único baseado no timestamp
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
