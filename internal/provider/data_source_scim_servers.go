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

// ScimServerItem representa um servidor SCIM do JumpCloud no data source
type ScimServerItem struct {
	ID              string `json:"_id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	Type            string `json:"type"`
	URL             string `json:"url"`
	Enabled         bool   `json:"enabled"`
	AuthType        string `json:"authType"`
	MappingSchemaID string `json:"mappingSchemaId,omitempty"`
	ScheduleType    string `json:"scheduleType,omitempty"`
	SyncInterval    int    `json:"syncInterval,omitempty"`
	Status          string `json:"status,omitempty"`
	OrgID           string `json:"orgId,omitempty"`
	Created         string `json:"created"`
	Updated         string `json:"updated"`
	LastSync        string `json:"lastSync,omitempty"`
}

// ScimServersResponse representa a resposta da API para listagem de servidores SCIM
type ScimServersResponse struct {
	Results    []ScimServerItem `json:"results"`
	TotalCount int              `json:"totalCount"`
}

func dataSourceScimServers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScimServersRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar por nome do servidor",
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar por tipo do servidor (saas, identity_provider, custom)",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Filtrar por status de habilitado",
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar por status (active, error, syncing)",
			},
			"auth_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar por tipo de autenticação (bearer, basic, oauth2)",
			},
			"search": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar servidores por texto em nome ou descrição",
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     100,
				Description: "Número máximo de servidores a serem retornados",
			},
			"skip": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Número de servidores a serem ignorados",
			},
			"sort": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "name",
				Description: "Campo para ordenação dos resultados",
			},
			"sort_dir": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "asc",
				Description: "Direção da ordenação (asc ou desc)",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"servers": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Lista de servidores SCIM encontrados",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do servidor SCIM",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome do servidor SCIM",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Descrição do servidor SCIM",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo do servidor (saas, identity_provider, custom)",
						},
						"url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL base para o endpoint SCIM",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indica se o servidor SCIM está habilitado",
						},
						"auth_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo de autenticação (bearer, basic, oauth2)",
						},
						"mapping_schema_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do esquema de mapeamento SCIM associado",
						},
						"schedule_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo de agendamento de sincronização (manual, scheduled)",
						},
						"sync_interval": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Intervalo de sincronização em minutos",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status atual do servidor SCIM (active, error, syncing)",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da organização",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data de criação do servidor SCIM",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data da última atualização do servidor SCIM",
						},
						"last_sync": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data da última sincronização",
						},
					},
				},
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de servidores encontrados",
			},
		},
	}
}

func dataSourceScimServersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Construir parâmetros de consulta
	queryParams := constructScimServersQueryParams(d)

	// Construir URL com parâmetros
	url := fmt.Sprintf("/api/v2/scim/servers?%s", queryParams)

	// Buscar servidores via API
	tflog.Debug(ctx, fmt.Sprintf("Listando servidores SCIM com parâmetros: %s", queryParams))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao listar servidores SCIM: %v", err))
	}

	// Deserializar resposta
	var serversResp ScimServersResponse
	if err := json.Unmarshal(resp, &serversResp); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Converter servidores para formato Terraform
	tfServers := flattenScimServers(serversResp.Results)
	if err := d.Set("servers", tfServers); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir servers: %v", err))
	}

	d.Set("total", serversResp.TotalCount)

	// Gerar ID único para o data source
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

// Função auxiliar para construir os parâmetros de consulta
func constructScimServersQueryParams(d *schema.ResourceData) string {
	params := ""

	// Adicionar filtros
	if v, ok := d.GetOk("name"); ok {
		params += fmt.Sprintf("name=%s&", v.(string))
	}

	if v, ok := d.GetOk("type"); ok {
		params += fmt.Sprintf("type=%s&", v.(string))
	}

	if v, ok := d.GetOk("enabled"); ok {
		params += fmt.Sprintf("enabled=%t&", v.(bool))
	}

	if v, ok := d.GetOk("status"); ok {
		params += fmt.Sprintf("status=%s&", v.(string))
	}

	if v, ok := d.GetOk("auth_type"); ok {
		params += fmt.Sprintf("authType=%s&", v.(string))
	}

	if v, ok := d.GetOk("search"); ok {
		params += fmt.Sprintf("search=%s&", v.(string))
	}

	// Adicionar parâmetros de paginação e ordenação
	params += fmt.Sprintf("limit=%d&", d.Get("limit").(int))
	params += fmt.Sprintf("skip=%d&", d.Get("skip").(int))
	params += fmt.Sprintf("sort=%s&", d.Get("sort").(string))
	params += fmt.Sprintf("sort_dir=%s&", d.Get("sort_dir").(string))

	// Adicionar org_id se fornecido
	if v, ok := d.GetOk("org_id"); ok {
		params += fmt.Sprintf("orgId=%s&", v.(string))
	}

	// Remover último & se existir
	if len(params) > 0 && params[len(params)-1] == '&' {
		params = params[:len(params)-1]
	}

	return params
}

// Função auxiliar para converter servidores para formato adequado ao Terraform
func flattenScimServers(servers []ScimServerItem) []map[string]interface{} {
	result := make([]map[string]interface{}, len(servers))

	for i, server := range servers {
		serverMap := map[string]interface{}{
			"id":            server.ID,
			"name":          server.Name,
			"description":   server.Description,
			"type":          server.Type,
			"url":           server.URL,
			"enabled":       server.Enabled,
			"auth_type":     server.AuthType,
			"schedule_type": server.ScheduleType,
			"sync_interval": server.SyncInterval,
			"status":        server.Status,
			"created":       server.Created,
			"updated":       server.Updated,
		}

		// Campos opcionais
		if server.MappingSchemaID != "" {
			serverMap["mapping_schema_id"] = server.MappingSchemaID
		}

		if server.LastSync != "" {
			serverMap["last_sync"] = server.LastSync
		}

		if server.OrgID != "" {
			serverMap["org_id"] = server.OrgID
		}

		result[i] = serverMap
	}

	return result
}
