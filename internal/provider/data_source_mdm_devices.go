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

// MDMDevice representa um dispositivo gerenciado por MDM no JumpCloud
type MDMDevice struct {
	ID                  string            `json:"_id"`
	SystemID            string            `json:"systemId,omitempty"`
	DisplayName         string            `json:"displayName"`
	Model               string            `json:"model,omitempty"`
	Manufacturer        string            `json:"manufacturer,omitempty"`
	OsVersion           string            `json:"osVersion,omitempty"`
	Platform            string            `json:"platform"` // ios, android, windows
	SerialNumber        string            `json:"serialNumber,omitempty"`
	UDID                string            `json:"udid,omitempty"`
	IMEI                string            `json:"imei,omitempty"`
	LastCheckIn         string            `json:"lastCheckIn,omitempty"`
	EnrollmentStatus    string            `json:"enrollmentStatus"`
	Managed             bool              `json:"managed"`
	Compliant           bool              `json:"compliant"`
	UserID              string            `json:"userId,omitempty"`
	UserName            string            `json:"userName,omitempty"`
	LastEnrolledDate    string            `json:"lastEnrolledDate,omitempty"`
	LastEnrollmentType  string            `json:"lastEnrollmentType,omitempty"`
	DeviceOwnershipType string            `json:"deviceOwnershipType,omitempty"` // personal, corporate
	Tags                []string          `json:"tags,omitempty"`
	Attributes          map[string]string `json:"attributes,omitempty"`
	OrgID               string            `json:"orgId,omitempty"`
	Created             string            `json:"created,omitempty"`
	Updated             string            `json:"updated,omitempty"`
}

// MDMDevicesResponse representa a resposta da API para a consulta de dispositivos MDM
type MDMDevicesResponse struct {
	Results     []MDMDevice `json:"results"`
	TotalCount  int         `json:"totalCount"`
	NextPageURL string      `json:"nextPageUrl,omitempty"`
}

func dataSourceMDMDevices() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMDMDevicesRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"platform": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"ios", "android", "windows", "all"}, false),
							Description:  "Filtrar por plataforma do dispositivo (ios, android, windows, all)",
							Default:      "all",
						},
						"enrollment_status": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"enrolled", "pending", "removed", "all"}, false),
							Description:  "Filtrar por status de registro (enrolled, pending, removed, all)",
							Default:      "all",
						},
						"managed": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Filtrar por estado de gerenciamento",
						},
						"compliant": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Filtrar por estado de conformidade",
						},
						"user_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filtrar por ID do usuário",
						},
						"device_ownership_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"personal", "corporate", "all"}, false),
							Description:  "Filtrar por tipo de propriedade do dispositivo (personal, corporate, all)",
							Default:      "all",
						},
						"search": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Termo de busca para filtrar dispositivos (nome, modelo, número de série, etc.)",
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
							ValidateFunc: validation.StringInSlice([]string{"displayName", "platform", "model", "lastCheckIn", "enrollmentStatus", "created"}, false),
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
				Description: "Número máximo de dispositivos a retornar",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambiente multi-tenant",
			},
			"devices": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do dispositivo MDM",
						},
						"system_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do sistema associado ao dispositivo",
						},
						"display_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome de exibição do dispositivo",
						},
						"model": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Modelo do dispositivo",
						},
						"manufacturer": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Fabricante do dispositivo",
						},
						"os_version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Versão do sistema operacional",
						},
						"platform": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Plataforma do dispositivo (ios, android, windows)",
						},
						"serial_number": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Número de série do dispositivo",
						},
						"udid": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "UDID do dispositivo (principalmente para iOS)",
						},
						"imei": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IMEI do dispositivo",
						},
						"last_check_in": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data e hora do último check-in",
						},
						"enrollment_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status do registro (enrolled, pending, removed)",
						},
						"managed": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Se o dispositivo está gerenciado",
						},
						"compliant": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Se o dispositivo está em conformidade com as políticas",
						},
						"user_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do usuário associado ao dispositivo",
						},
						"user_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome do usuário associado ao dispositivo",
						},
						"last_enrolled_date": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data e hora do último registro",
						},
						"last_enrollment_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo do último registro",
						},
						"device_ownership_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo de propriedade do dispositivo (personal, corporate)",
						},
						"tags": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Tags associadas ao dispositivo",
						},
						"attributes": {
							Type:        schema.TypeMap,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Atributos personalizados do dispositivo",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data e hora de criação do registro do dispositivo",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data e hora da última atualização do registro do dispositivo",
						},
					},
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de dispositivos que correspondem aos filtros",
			},
		},
	}
}

func dataSourceMDMDevicesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir URL base para a requisição
	url := "/api/v2/mdm/devices"
	queryParams := ""

	// Aplicar filtros
	if filterList, ok := d.GetOk("filter"); ok && len(filterList.([]interface{})) > 0 {
		filter := filterList.([]interface{})[0].(map[string]interface{})

		if platform, ok := filter["platform"]; ok && platform.(string) != "all" {
			queryParams += fmt.Sprintf("&platform=%s", platform.(string))
		}

		if status, ok := filter["enrollment_status"]; ok && status.(string) != "all" {
			queryParams += fmt.Sprintf("&enrollmentStatus=%s", status.(string))
		}

		if managed, ok := filter["managed"]; ok {
			queryParams += fmt.Sprintf("&managed=%t", managed.(bool))
		}

		if compliant, ok := filter["compliant"]; ok {
			queryParams += fmt.Sprintf("&compliant=%t", compliant.(bool))
		}

		if userID, ok := filter["user_id"]; ok && userID.(string) != "" {
			queryParams += fmt.Sprintf("&userId=%s", userID.(string))
		}

		if ownershipType, ok := filter["device_ownership_type"]; ok && ownershipType.(string) != "all" {
			queryParams += fmt.Sprintf("&deviceOwnershipType=%s", ownershipType.(string))
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
	tflog.Debug(ctx, fmt.Sprintf("Consultando dispositivos MDM: %s%s", url, queryParams))
	resp, err := c.DoRequest(http.MethodGet, url+queryParams, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar dispositivos MDM: %v", err))
	}

	// Deserializar resposta
	var response MDMDevicesResponse
	if err := json.Unmarshal(resp, &response); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Converter lista de dispositivos para o formato do Terraform
	devices := make([]map[string]interface{}, 0, len(response.Results))
	for _, device := range response.Results {
		deviceMap := map[string]interface{}{
			"id":                device.ID,
			"display_name":      device.DisplayName,
			"platform":          device.Platform,
			"enrollment_status": device.EnrollmentStatus,
			"managed":           device.Managed,
			"compliant":         device.Compliant,
			"created":           device.Created,
			"updated":           device.Updated,
		}

		// Campos opcionais
		if device.SystemID != "" {
			deviceMap["system_id"] = device.SystemID
		}
		if device.Model != "" {
			deviceMap["model"] = device.Model
		}
		if device.Manufacturer != "" {
			deviceMap["manufacturer"] = device.Manufacturer
		}
		if device.OsVersion != "" {
			deviceMap["os_version"] = device.OsVersion
		}
		if device.SerialNumber != "" {
			deviceMap["serial_number"] = device.SerialNumber
		}
		if device.UDID != "" {
			deviceMap["udid"] = device.UDID
		}
		if device.IMEI != "" {
			deviceMap["imei"] = device.IMEI
		}
		if device.LastCheckIn != "" {
			deviceMap["last_check_in"] = device.LastCheckIn
		}
		if device.UserID != "" {
			deviceMap["user_id"] = device.UserID
		}
		if device.UserName != "" {
			deviceMap["user_name"] = device.UserName
		}
		if device.LastEnrolledDate != "" {
			deviceMap["last_enrolled_date"] = device.LastEnrolledDate
		}
		if device.LastEnrollmentType != "" {
			deviceMap["last_enrollment_type"] = device.LastEnrollmentType
		}
		if device.DeviceOwnershipType != "" {
			deviceMap["device_ownership_type"] = device.DeviceOwnershipType
		}
		if len(device.Tags) > 0 {
			deviceMap["tags"] = device.Tags
		}
		if len(device.Attributes) > 0 {
			deviceMap["attributes"] = device.Attributes
		}

		devices = append(devices, deviceMap)
	}

	// Definir valores no state
	if err := d.Set("devices", devices); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir dispositivos no state: %v", err))
	}

	if err := d.Set("total_count", response.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir total_count no state: %v", err))
	}

	// Definir ID único para o data source (baseado no timestamp atual)
	d.SetId(fmt.Sprintf("mdm-devices-%d", time.Now().Unix()))

	return diags
}
