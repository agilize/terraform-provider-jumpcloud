package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// ScimIntegration representa uma integração SCIM no JumpCloud
type ScimIntegration struct {
	ID             string                 `json:"_id,omitempty"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	Type           string                 `json:"type"` // saas, identity_provider, custom
	ServerID       string                 `json:"serverId"`
	Status         string                 `json:"status,omitempty"` // active, pending, error
	Enabled        bool                   `json:"enabled"`
	Settings       map[string]interface{} `json:"settings,omitempty"`
	SyncSchedule   string                 `json:"syncSchedule,omitempty"` // manual, daily, hourly, etc
	SyncInterval   int                    `json:"syncInterval,omitempty"` // em minutos
	LastSyncTime   string                 `json:"lastSyncTime,omitempty"`
	LastSyncStatus string                 `json:"lastSyncStatus,omitempty"`
	MappingIDs     []string               `json:"mappingIds,omitempty"`
	OrgID          string                 `json:"orgId,omitempty"`
	Created        string                 `json:"created,omitempty"`
	Updated        string                 `json:"updated,omitempty"`
}

func resourceScimIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScimIntegrationCreate,
		ReadContext:   resourceScimIntegrationRead,
		UpdateContext: resourceScimIntegrationUpdate,
		DeleteContext: resourceScimIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
				Description:  "Nome da integração SCIM",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição da integração SCIM",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"saas", "identity_provider", "custom",
				}, false),
				Description: "Tipo da integração SCIM (saas, identity_provider, custom)",
			},
			"server_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do servidor SCIM associado à integração",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indica se a integração SCIM está ativada",
			},
			"settings": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJSONDiffs,
				Description:      "Configurações específicas da integração em formato JSON",
			},
			"sync_schedule": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "manual",
				ValidateFunc: validation.StringInSlice([]string{
					"manual", "hourly", "daily", "weekly", "custom",
				}, false),
				Description: "Agendamento de sincronização (manual, hourly, daily, weekly, custom)",
			},
			"sync_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1440, // 24 horas em minutos
				ValidateFunc: validation.IntAtLeast(15),
				Description:  "Intervalo de sincronização em minutos (quando sync_schedule é 'custom')",
			},
			"mapping_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "IDs dos mapeamentos de atributos associados",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status atual da integração (active, pending, error)",
			},
			"last_sync_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data e hora da última sincronização",
			},
			"last_sync_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status da última sincronização",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da integração",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da integração",
			},
		},
	}
}

// suppressEquivalentJSONDiffs suprime diferenças em strings JSON que são estruturalmente equivalentes
func suppressEquivalentJSONDiffs(k, old, new string, d *schema.ResourceData) bool {
	// Se ambos estiverem vazios, são iguais
	if old == "" && new == "" {
		return true
	}

	// Se apenas um estiver vazio, não são iguais
	if old == "" || new == "" {
		return false
	}

	var oldObj, newObj interface{}
	if err := json.Unmarshal([]byte(old), &oldObj); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &newObj); err != nil {
		return false
	}

	// Comparar a representação JSON normalizada
	oldNormalized, err := json.Marshal(oldObj)
	if err != nil {
		return false
	}
	newNormalized, err := json.Marshal(newObj)
	if err != nil {
		return false
	}

	return string(oldNormalized) == string(newNormalized)
}

func resourceScimIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir objeto ScimIntegration a partir dos dados do terraform
	integration := &ScimIntegration{
		Name:         d.Get("name").(string),
		Type:         d.Get("type").(string),
		ServerID:     d.Get("server_id").(string),
		Enabled:      d.Get("enabled").(bool),
		SyncSchedule: d.Get("sync_schedule").(string),
		SyncInterval: d.Get("sync_interval").(int),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		integration.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		integration.OrgID = v.(string)
	}

	// Processar configurações (JSON)
	if v, ok := d.GetOk("settings"); ok {
		settingsJSON := v.(string)
		var settings map[string]interface{}
		if err := json.Unmarshal([]byte(settingsJSON), &settings); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar configurações: %v", err))
		}
		integration.Settings = settings
	}

	// Processar IDs de mapeamento
	if v, ok := d.GetOk("mapping_ids"); ok {
		mappingList := v.([]interface{})
		mappings := make([]string, len(mappingList))
		for i, meta := range mappingList {
			mappings[i] = meta.(string)
		}
		integration.MappingIDs = mappings
	}

	// Serializar para JSON
	reqBody, err := json.Marshal(integration)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar integração SCIM: %v", err))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/scim/servers/%s/integrations", integration.ServerID)
	if integration.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, integration.OrgID)
	}

	// Fazer requisição para criar integração
	tflog.Debug(ctx, fmt.Sprintf("Criando integração SCIM para servidor: %s", integration.ServerID))
	resp, err := c.DoRequest(http.MethodPost, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar integração SCIM: %v", err))
	}

	// Deserializar resposta
	var createdIntegration ScimIntegration
	if err := json.Unmarshal(resp, &createdIntegration); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir ID no state
	d.SetId(createdIntegration.ID)

	// Ler o recurso para atualizar o state com todos os campos computados
	return resourceScimIntegrationRead(ctx, d, meta)
}

func resourceScimIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID da integração
	integrationID := d.Id()

	// Obter o ID do servidor do state
	var serverID string
	if v, ok := d.GetOk("server_id"); ok {
		serverID = v.(string)
	} else {
		// Se não tivermos o server_id no state (possivelmente durante importação),
		// precisamos buscar a integração pelo ID para descobrir o server_id
		url := fmt.Sprintf("/api/v2/scim/integrations/%s", integrationID)
		if v, ok := d.GetOk("org_id"); ok {
			url = fmt.Sprintf("%s?orgId=%s", url, v.(string))
		}

		resp, err := c.DoRequest(http.MethodGet, url, nil)
		if err != nil {
			if err.Error() == "Status Code: 404" {
				d.SetId("")
				return diags
			}
			return diag.FromErr(fmt.Errorf("erro ao buscar integração SCIM: %v", err))
		}

		var integration ScimIntegration
		if err := json.Unmarshal(resp, &integration); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
		}

		serverID = integration.ServerID
		d.Set("server_id", serverID)
	}

	// Obter parâmetro orgId se disponível
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/scim/servers/%s/integrations/%s%s", serverID, integrationID, orgIDParam)

	// Fazer requisição para ler integração
	tflog.Debug(ctx, fmt.Sprintf("Lendo integração SCIM: %s", integrationID))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		// Se o recurso não for encontrado, remover do state
		if err.Error() == "Status Code: 404" {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler integração SCIM: %v", err))
	}

	// Deserializar resposta
	var integration ScimIntegration
	if err := json.Unmarshal(resp, &integration); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Mapear valores para o schema
	d.Set("name", integration.Name)
	d.Set("description", integration.Description)
	d.Set("type", integration.Type)
	d.Set("server_id", integration.ServerID)
	d.Set("enabled", integration.Enabled)
	d.Set("status", integration.Status)
	d.Set("sync_schedule", integration.SyncSchedule)
	d.Set("sync_interval", integration.SyncInterval)
	d.Set("last_sync_time", integration.LastSyncTime)
	d.Set("last_sync_status", integration.LastSyncStatus)
	d.Set("created", integration.Created)
	d.Set("updated", integration.Updated)

	// Serializar configurações para JSON
	if integration.Settings != nil {
		settingsJSON, err := json.Marshal(integration.Settings)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao serializar configurações: %v", err))
		}
		d.Set("settings", string(settingsJSON))
	}

	// Definir IDs de mapeamento
	if err := d.Set("mapping_ids", integration.MappingIDs); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir mapping_ids: %v", err))
	}

	// Definir OrgID se presente
	if integration.OrgID != "" {
		d.Set("org_id", integration.OrgID)
	}

	return diags
}

func resourceScimIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID da integração
	integrationID := d.Id()

	// Construir objeto ScimIntegration a partir dos dados do terraform
	integration := &ScimIntegration{
		ID:           integrationID,
		Name:         d.Get("name").(string),
		Type:         d.Get("type").(string),
		ServerID:     d.Get("server_id").(string),
		Enabled:      d.Get("enabled").(bool),
		SyncSchedule: d.Get("sync_schedule").(string),
		SyncInterval: d.Get("sync_interval").(int),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		integration.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		integration.OrgID = v.(string)
	}

	// Processar configurações (JSON)
	if v, ok := d.GetOk("settings"); ok {
		settingsJSON := v.(string)
		var settings map[string]interface{}
		if err := json.Unmarshal([]byte(settingsJSON), &settings); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar configurações: %v", err))
		}
		integration.Settings = settings
	}

	// Processar IDs de mapeamento
	if v, ok := d.GetOk("mapping_ids"); ok {
		mappingList := v.([]interface{})
		mappings := make([]string, len(mappingList))
		for i, meta := range mappingList {
			mappings[i] = meta.(string)
		}
		integration.MappingIDs = mappings
	}

	// Serializar para JSON
	reqBody, err := json.Marshal(integration)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar integração SCIM: %v", err))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/scim/servers/%s/integrations/%s", integration.ServerID, integrationID)
	if integration.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, integration.OrgID)
	}

	// Fazer requisição para atualizar integração
	tflog.Debug(ctx, fmt.Sprintf("Atualizando integração SCIM: %s", integrationID))
	_, err = c.DoRequest(http.MethodPut, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar integração SCIM: %v", err))
	}

	// Ler o recurso para atualizar o state com todos os campos computados
	return resourceScimIntegrationRead(ctx, d, meta)
}

func resourceScimIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID da integração
	integrationID := d.Id()

	// Obter ID do servidor
	serverID := d.Get("server_id").(string)

	// Obter parâmetro orgId se disponível
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/scim/servers/%s/integrations/%s%s", serverID, integrationID, orgIDParam)

	// Fazer requisição para excluir integração
	tflog.Debug(ctx, fmt.Sprintf("Excluindo integração SCIM: %s", integrationID))
	_, err := c.DoRequest(http.MethodDelete, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao excluir integração SCIM: %v", err))
	}

	// Remover ID do state
	d.SetId("")

	return diags
}
