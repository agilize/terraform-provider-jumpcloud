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

// DirectoryInsightsEventsRequest representa os parâmetros de busca de eventos
type DirectoryInsightsEventsRequest struct {
	StartTime      time.Time              `json:"startTime"`
	EndTime        time.Time              `json:"endTime"`
	Limit          int                    `json:"limit,omitempty"`
	Skip           int                    `json:"skip,omitempty"`
	SearchTermAnd  []string               `json:"searchTermAnd,omitempty"`
	SearchTermOr   []string               `json:"searchTermOr,omitempty"`
	Service        []string               `json:"service,omitempty"`
	EventType      []string               `json:"eventType,omitempty"`
	SortOrder      string                 `json:"sortOrder,omitempty"`
	InitiatedBy    map[string]interface{} `json:"initiatedBy,omitempty"`
	Resource       map[string]interface{} `json:"resource,omitempty"`
	TimeRange      string                 `json:"timeRange,omitempty"`
	UseDefaultSort bool                   `json:"useDefaultSort,omitempty"`
}

// DirectoryInsightsEvent representa um evento retornado pelo Directory Insights
type DirectoryInsightsEvent struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Timestamp    string                 `json:"timestamp"`
	Service      string                 `json:"service"`
	ClientIP     string                 `json:"client_ip,omitempty"`
	Resource     map[string]interface{} `json:"resource,omitempty"`
	Success      bool                   `json:"success"`
	Message      string                 `json:"message,omitempty"`
	GeoIP        map[string]interface{} `json:"geoip,omitempty"`
	InitiatedBy  map[string]interface{} `json:"initiated_by,omitempty"`
	Changes      map[string]interface{} `json:"changes,omitempty"`
	RawEventType string                 `json:"raw_event_type,omitempty"`
	OrgId        string                 `json:"organization,omitempty"`
}

// DirectoryInsightsEventsResponse representa a resposta da API de eventos
type DirectoryInsightsEventsResponse struct {
	Results     []DirectoryInsightsEvent `json:"results"`
	TotalCount  int                      `json:"totalCount"`
	HasMore     bool                     `json:"hasMore"`
	NextOffset  int                      `json:"nextOffset,omitempty"`
	NextPageURL string                   `json:"nextPageUrl,omitempty"`
}

func dataSourceDirectoryInsightsEvents() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDirectoryInsightsEventsRead,
		Schema: map[string]*schema.Schema{
			"start_time": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsRFC3339Time,
				Description:  "Horário de início para busca de eventos (formato RFC3339)",
			},
			"end_time": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsRFC3339Time,
				Description:  "Horário de fim para busca de eventos (formato RFC3339)",
			},
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 1000),
				Description:  "Número máximo de eventos a serem retornados (1-1000)",
			},
			"skip": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Número de eventos a serem pulados (paginação)",
			},
			"search_term_and": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Lista de termos de busca que devem todos aparecer nos eventos (condição AND)",
			},
			"search_term_or": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Lista de termos de busca onde pelo menos um deve aparecer nos eventos (condição OR)",
			},
			"service": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Lista de serviços para filtrar eventos (ex: directory, radius, sso)",
			},
			"event_type": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Lista de tipos de eventos para filtrar",
			},
			"sort_order": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "DESC",
				ValidateFunc: validation.StringInSlice([]string{"ASC", "DESC"}, false),
				Description:  "Ordenação dos eventos: ASC (mais antigos primeiro) ou DESC (mais recentes primeiro)",
			},
			"user_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar eventos iniciados por um usuário específico",
			},
			"admin_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar eventos iniciados por um administrador específico",
			},
			"resource_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar eventos relacionados a um recurso específico",
			},
			"resource_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar eventos por tipo de recurso",
			},
			"time_range": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"1d", "7d", "30d", "custom"}, false),
				Description:  "Período predefinido para busca: 1d (1 dia), 7d (7 dias), 30d (30 dias) ou custom (personalizado)",
			},
			"use_default_sort": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Se deve usar a ordenação padrão (por timestamp)",
			},
			"events": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID único do evento",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo do evento",
						},
						"timestamp": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp do evento",
						},
						"service": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Serviço que gerou o evento",
						},
						"client_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP do cliente que iniciou o evento",
						},
						"success": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Se o evento foi bem-sucedido",
						},
						"message": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Mensagem descritiva do evento",
						},
						"raw_event_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo de evento bruto",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da organização",
						},
						"resource_json": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Informações do recurso como JSON",
						},
						"initiated_by_json": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Informações de quem iniciou o evento como JSON",
						},
						"changes_json": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Informações das alterações como JSON",
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
				Description: "Número total de eventos que correspondem aos critérios",
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Se há mais eventos disponíveis além dos retornados",
			},
			"next_offset": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Offset para a próxima página de resultados",
			},
		},
	}
}

func dataSourceDirectoryInsightsEventsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir request de busca de eventos
	req := &DirectoryInsightsEventsRequest{
		Limit:          d.Get("limit").(int),
		Skip:           d.Get("skip").(int),
		SortOrder:      d.Get("sort_order").(string),
		UseDefaultSort: d.Get("use_default_sort").(bool),
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
	if v, ok := d.GetOk("search_term_and"); ok {
		terms := v.([]interface{})
		req.SearchTermAnd = make([]string, len(terms))
		for i, term := range terms {
			req.SearchTermAnd[i] = term.(string)
		}
	}

	if v, ok := d.GetOk("search_term_or"); ok {
		terms := v.([]interface{})
		req.SearchTermOr = make([]string, len(terms))
		for i, term := range terms {
			req.SearchTermOr[i] = term.(string)
		}
	}

	if v, ok := d.GetOk("service"); ok {
		services := v.([]interface{})
		req.Service = make([]string, len(services))
		for i, service := range services {
			req.Service[i] = service.(string)
		}
	}

	if v, ok := d.GetOk("event_type"); ok {
		types := v.([]interface{})
		req.EventType = make([]string, len(types))
		for i, t := range types {
			req.EventType[i] = t.(string)
		}
	}

	if v, ok := d.GetOk("time_range"); ok {
		req.TimeRange = v.(string)
	}

	// Processar filtros de usuário/admin/recurso
	initiatedBy := make(map[string]interface{})

	if v, ok := d.GetOk("user_id"); ok {
		initiatedBy["user_id"] = v.(string)
	}

	if v, ok := d.GetOk("admin_id"); ok {
		initiatedBy["admin_id"] = v.(string)
	}

	if len(initiatedBy) > 0 {
		req.InitiatedBy = initiatedBy
	}

	resource := make(map[string]interface{})

	if v, ok := d.GetOk("resource_id"); ok {
		resource["id"] = v.(string)
	}

	if v, ok := d.GetOk("resource_type"); ok {
		resource["type"] = v.(string)
	}

	if len(resource) > 0 {
		req.Resource = resource
	}

	// Serializar para JSON
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar requisição: %v", err))
	}

	// Buscar eventos via API
	tflog.Debug(ctx, "Buscando eventos no Directory Insights")
	resp, err := c.DoRequest(http.MethodPost, "/insights/directory/v1/events", reqJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao buscar eventos no Directory Insights: %v", err))
	}

	// Deserializar resposta
	var eventsResp DirectoryInsightsEventsResponse
	if err := json.Unmarshal(resp, &eventsResp); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Processar eventos e definir no state
	events := make([]map[string]interface{}, len(eventsResp.Results))
	for i, event := range eventsResp.Results {
		// Serializar campos complexos para JSON
		resourceJSON, _ := json.Marshal(event.Resource)
		initiatedByJSON, _ := json.Marshal(event.InitiatedBy)
		changesJSON, _ := json.Marshal(event.Changes)
		geoIPJSON, _ := json.Marshal(event.GeoIP)

		events[i] = map[string]interface{}{
			"id":                event.ID,
			"type":              event.Type,
			"timestamp":         event.Timestamp,
			"service":           event.Service,
			"client_ip":         event.ClientIP,
			"success":           event.Success,
			"message":           event.Message,
			"raw_event_type":    event.RawEventType,
			"org_id":            event.OrgId,
			"resource_json":     string(resourceJSON),
			"initiated_by_json": string(initiatedByJSON),
			"changes_json":      string(changesJSON),
			"geoip_json":        string(geoIPJSON),
		}
	}

	// Atualizar o state
	d.SetId(time.Now().Format(time.RFC3339)) // ID único para o data source
	d.Set("events", events)
	d.Set("total_count", eventsResp.TotalCount)
	d.Set("has_more", eventsResp.HasMore)
	d.Set("next_offset", eventsResp.NextOffset)

	return diag.Diagnostics{}
}
