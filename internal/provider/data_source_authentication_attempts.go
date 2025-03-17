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

// AuthAttemptRequest representa os parâmetros de busca de tentativas de autenticação
type AuthAttemptRequest struct {
	StartTime     time.Time              `json:"startTime"`
	EndTime       time.Time              `json:"endTime"`
	Limit         int                    `json:"limit,omitempty"`
	Skip          int                    `json:"skip,omitempty"`
	SearchTerm    string                 `json:"searchTerm,omitempty"`
	Service       []string               `json:"service,omitempty"`
	Success       *bool                  `json:"success,omitempty"`
	SortOrder     string                 `json:"sortOrder,omitempty"`
	UserID        string                 `json:"userId,omitempty"`
	SystemID      string                 `json:"systemId,omitempty"`
	ApplicationID string                 `json:"applicationId,omitempty"`
	IPAddress     string                 `json:"ipAddress,omitempty"`
	GeoIP         map[string]interface{} `json:"geoip,omitempty"`
	TimeRange     string                 `json:"timeRange,omitempty"`
}

// AuthAttempt representa uma tentativa de autenticação no JumpCloud
type AuthAttempt struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Timestamp     string                 `json:"timestamp"`
	Service       string                 `json:"service"`
	ClientIP      string                 `json:"client_ip,omitempty"`
	Success       bool                   `json:"success"`
	Message       string                 `json:"message,omitempty"`
	GeoIP         map[string]interface{} `json:"geoip,omitempty"`
	RawEventType  string                 `json:"raw_event_type,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
	Username      string                 `json:"username,omitempty"`
	SystemID      string                 `json:"system_id,omitempty"`
	SystemName    string                 `json:"system_name,omitempty"`
	ApplicationID string                 `json:"application_id,omitempty"`
	AppName       string                 `json:"app_name,omitempty"`
	MFAType       string                 `json:"mfa_type,omitempty"`
	OrgID         string                 `json:"organization,omitempty"`
}

// AuthAttemptsResponse representa a resposta da API de tentativas de autenticação
type AuthAttemptsResponse struct {
	Results     []AuthAttempt `json:"results"`
	TotalCount  int           `json:"totalCount"`
	HasMore     bool          `json:"hasMore"`
	NextOffset  int           `json:"nextOffset,omitempty"`
	NextPageURL string        `json:"nextPageUrl,omitempty"`
}

func dataSourceAuthenticationAttempts() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAuthenticationAttemptsRead,
		Schema: map[string]*schema.Schema{
			"start_time": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsRFC3339Time,
				Description:  "Horário de início para busca de tentativas (formato RFC3339)",
			},
			"end_time": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsRFC3339Time,
				Description:  "Horário de fim para busca de tentativas (formato RFC3339)",
			},
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 1000),
				Description:  "Número máximo de tentativas a serem retornadas (1-1000)",
			},
			"skip": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Número de tentativas a serem puladas (paginação)",
			},
			"search_term": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Termo para busca em todos os campos",
			},
			"service": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Lista de serviços para filtrar tentativas (ex: radius, sso, ldap, system)",
			},
			"success": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Filtrar apenas tentativas bem-sucedidas (true) ou falhas (false)",
			},
			"sort_order": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "DESC",
				ValidateFunc: validation.StringInSlice([]string{"ASC", "DESC"}, false),
				Description:  "Ordenação das tentativas: ASC (mais antigas primeiro) ou DESC (mais recentes primeiro)",
			},
			"user_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar tentativas de um usuário específico",
			},
			"system_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar tentativas em um sistema específico",
			},
			"application_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar tentativas em uma aplicação específica",
			},
			"ip_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar tentativas por endereço IP",
			},
			"country_code": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar tentativas por código de país",
			},
			"time_range": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"1d", "7d", "30d", "custom"}, false),
				Description:  "Período predefinido para busca: 1d (1 dia), 7d (7 dias), 30d (30 dias) ou custom (personalizado)",
			},
			"attempts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID único da tentativa",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo da tentativa de autenticação",
						},
						"timestamp": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp da tentativa",
						},
						"service": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Serviço usado na tentativa (radius, sso, ldap, system, etc)",
						},
						"client_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP do cliente que tentou autenticação",
						},
						"success": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Se a tentativa foi bem-sucedida",
						},
						"message": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Mensagem descritiva da tentativa",
						},
						"raw_event_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo de evento bruto",
						},
						"user_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do usuário que tentou autenticação",
						},
						"username": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome do usuário que tentou autenticação",
						},
						"system_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do sistema onde a tentativa ocorreu",
						},
						"system_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome do sistema onde a tentativa ocorreu",
						},
						"application_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da aplicação onde a tentativa ocorreu",
						},
						"application_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome da aplicação onde a tentativa ocorreu",
						},
						"mfa_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo de MFA usado na tentativa (totp, duo, push, fido)",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da organização",
						},
						"geoip_json": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Informações de geolocalização do IP como JSON",
						},
					},
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de tentativas que correspondem aos critérios",
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Se há mais tentativas disponíveis além das retornadas",
			},
			"next_offset": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Offset para a próxima página de resultados",
			},
		},
	}
}

func dataSourceAuthenticationAttemptsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir request de busca de tentativas
	req := &AuthAttemptRequest{
		Limit:     d.Get("limit").(int),
		Skip:      d.Get("skip").(int),
		SortOrder: d.Get("sort_order").(string),
	}

	// Processar horários
	startTimeStr := d.Get("start_time").(string)
	endTimeStr := d.Get("end_time").(string)

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return diag.FromErr(fmt.Errorf("formato inválido para start_time: %v", err))
	}
	req.StartTime = startTime

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		return diag.FromErr(fmt.Errorf("formato inválido para end_time: %v", err))
	}
	req.EndTime = endTime

	// Processar campos opcionais
	if v, ok := d.GetOk("search_term"); ok {
		req.SearchTerm = v.(string)
	}

	if v, ok := d.GetOk("service"); ok {
		services := v.([]interface{})
		serviceList := make([]string, len(services))
		for i, s := range services {
			serviceList[i] = s.(string)
		}
		req.Service = serviceList
	}

	if v, ok := d.GetOk("success"); ok {
		success := v.(bool)
		req.Success = &success
	}

	if v, ok := d.GetOk("user_id"); ok {
		req.UserID = v.(string)
	}

	if v, ok := d.GetOk("system_id"); ok {
		req.SystemID = v.(string)
	}

	if v, ok := d.GetOk("application_id"); ok {
		req.ApplicationID = v.(string)
	}

	if v, ok := d.GetOk("ip_address"); ok {
		req.IPAddress = v.(string)
	}

	if v, ok := d.GetOk("country_code"); ok {
		geoIP := make(map[string]interface{})
		geoIP["country_code"] = v.(string)
		req.GeoIP = geoIP
	}

	if v, ok := d.GetOk("time_range"); ok {
		req.TimeRange = v.(string)
	}

	// Serializar para JSON
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar requisição: %v", err))
	}

	// Buscar tentativas via API
	tflog.Debug(ctx, "Buscando tentativas de autenticação")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/auth/attempts", reqJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao buscar tentativas de autenticação: %v", err))
	}

	// Deserializar resposta
	var attemptsResp AuthAttemptsResponse
	if err := json.Unmarshal(resp, &attemptsResp); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Processar tentativas e definir no state
	attempts := make([]map[string]interface{}, len(attemptsResp.Results))
	for i, attempt := range attemptsResp.Results {
		// Serializar campos complexos para JSON
		geoIPJSON, _ := json.Marshal(attempt.GeoIP)

		attempts[i] = map[string]interface{}{
			"id":               attempt.ID,
			"type":             attempt.Type,
			"timestamp":        attempt.Timestamp,
			"service":          attempt.Service,
			"client_ip":        attempt.ClientIP,
			"success":          attempt.Success,
			"message":          attempt.Message,
			"raw_event_type":   attempt.RawEventType,
			"user_id":          attempt.UserID,
			"username":         attempt.Username,
			"system_id":        attempt.SystemID,
			"system_name":      attempt.SystemName,
			"application_id":   attempt.ApplicationID,
			"application_name": attempt.AppName,
			"mfa_type":         attempt.MFAType,
			"org_id":           attempt.OrgID,
			"geoip_json":       string(geoIPJSON),
		}
	}

	// Atualizar o state
	d.SetId(time.Now().Format(time.RFC3339)) // ID único para o data source
	d.Set("attempts", attempts)
	d.Set("total_count", attemptsResp.TotalCount)
	d.Set("has_more", attemptsResp.HasMore)
	d.Set("next_offset", attemptsResp.NextOffset)

	return diag.Diagnostics{}
}
