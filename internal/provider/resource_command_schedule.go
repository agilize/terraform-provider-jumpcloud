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

// CommandSchedule representa um agendamento de comando no JumpCloud
type CommandSchedule struct {
	ID            string   `json:"_id,omitempty"`
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	CommandID     string   `json:"commandId"`
	Enabled       bool     `json:"enabled"`
	Schedule      string   `json:"schedule"`
	ScheduleType  string   `json:"scheduleType"` // cron, one-time
	Timezone      string   `json:"timezone,omitempty"`
	TargetSystems []string `json:"targetSystems,omitempty"`
	TargetGroups  []string `json:"targetGroups,omitempty"`
	OrgID         string   `json:"orgId,omitempty"`
	Created       string   `json:"created,omitempty"`
	Updated       string   `json:"updated,omitempty"`
}

// resourceCommandSchedule retorna o recurso para gerenciar agendamentos de comandos
func resourceCommandSchedule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCommandScheduleCreate,
		ReadContext:   resourceCommandScheduleRead,
		UpdateContext: resourceCommandScheduleUpdate,
		DeleteContext: resourceCommandScheduleDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome do agendamento",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição do agendamento",
			},
			"command_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID do comando a ser executado",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Se o agendamento está ativo",
			},
			"schedule": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Expressão cron ou timestamp para agendamento",
			},
			"schedule_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"cron", "one-time"}, false),
				Description:  "Tipo de agendamento: cron (recorrente) ou one-time (única vez)",
			},
			"timezone": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "UTC",
				Description: "Fuso horário para o agendamento (ex: America/Sao_Paulo)",
			},
			"target_systems": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IDs dos sistemas onde o comando será executado",
			},
			"target_groups": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IDs dos grupos onde o comando será executado",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação do agendamento",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização do agendamento",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// resourceCommandScheduleCreate cria um novo agendamento de comando
func resourceCommandScheduleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir agendamento
	schedule := &CommandSchedule{
		Name:         d.Get("name").(string),
		CommandID:    d.Get("command_id").(string),
		Enabled:      d.Get("enabled").(bool),
		Schedule:     d.Get("schedule").(string),
		ScheduleType: d.Get("schedule_type").(string),
		Timezone:     d.Get("timezone").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		schedule.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		schedule.OrgID = v.(string)
	}

	// Processar sistemas alvo
	if v, ok := d.GetOk("target_systems"); ok {
		systems := v.(*schema.Set).List()
		systemIDs := make([]string, len(systems))
		for i, s := range systems {
			systemIDs[i] = s.(string)
		}
		schedule.TargetSystems = systemIDs
	}

	// Processar grupos alvo
	if v, ok := d.GetOk("target_groups"); ok {
		groups := v.(*schema.Set).List()
		groupIDs := make([]string, len(groups))
		for i, g := range groups {
			groupIDs[i] = g.(string)
		}
		schedule.TargetGroups = groupIDs
	}

	// Verificar se pelo menos um alvo foi especificado
	if len(schedule.TargetSystems) == 0 && len(schedule.TargetGroups) == 0 {
		return diag.FromErr(fmt.Errorf("pelo menos um sistema alvo ou grupo alvo deve ser especificado"))
	}

	// Serializar para JSON
	scheduleJSON, err := json.Marshal(schedule)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar agendamento de comando: %v", err))
	}

	// Criar agendamento via API
	tflog.Debug(ctx, fmt.Sprintf("Criando agendamento de comando: %s", schedule.Name))
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/command/schedules", scheduleJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar agendamento de comando: %v", err))
	}

	// Deserializar resposta
	var createdSchedule CommandSchedule
	if err := json.Unmarshal(resp, &createdSchedule); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdSchedule.ID == "" {
		return diag.FromErr(fmt.Errorf("agendamento de comando criado sem ID"))
	}

	d.SetId(createdSchedule.ID)
	return resourceCommandScheduleRead(ctx, d, meta)
}

// resourceCommandScheduleRead lê os detalhes de um agendamento de comando
func resourceCommandScheduleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do agendamento de comando não fornecido"))
	}

	// Buscar agendamento via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo agendamento de comando com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/command/schedules/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Agendamento de comando %s não encontrado, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler agendamento de comando: %v", err))
	}

	// Deserializar resposta
	var schedule CommandSchedule
	if err := json.Unmarshal(resp, &schedule); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("name", schedule.Name)
	d.Set("description", schedule.Description)
	d.Set("command_id", schedule.CommandID)
	d.Set("enabled", schedule.Enabled)
	d.Set("schedule", schedule.Schedule)
	d.Set("schedule_type", schedule.ScheduleType)
	d.Set("timezone", schedule.Timezone)
	d.Set("created", schedule.Created)
	d.Set("updated", schedule.Updated)

	if schedule.OrgID != "" {
		d.Set("org_id", schedule.OrgID)
	}

	if schedule.TargetSystems != nil {
		d.Set("target_systems", schedule.TargetSystems)
	}

	if schedule.TargetGroups != nil {
		d.Set("target_groups", schedule.TargetGroups)
	}

	return diags
}

// resourceCommandScheduleUpdate atualiza um agendamento de comando existente
func resourceCommandScheduleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do agendamento de comando não fornecido"))
	}

	// Construir agendamento atualizado
	schedule := &CommandSchedule{
		ID:           id,
		Name:         d.Get("name").(string),
		CommandID:    d.Get("command_id").(string),
		Enabled:      d.Get("enabled").(bool),
		Schedule:     d.Get("schedule").(string),
		ScheduleType: d.Get("schedule_type").(string),
		Timezone:     d.Get("timezone").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		schedule.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		schedule.OrgID = v.(string)
	}

	// Processar sistemas alvo
	if v, ok := d.GetOk("target_systems"); ok {
		systems := v.(*schema.Set).List()
		systemIDs := make([]string, len(systems))
		for i, s := range systems {
			systemIDs[i] = s.(string)
		}
		schedule.TargetSystems = systemIDs
	}

	// Processar grupos alvo
	if v, ok := d.GetOk("target_groups"); ok {
		groups := v.(*schema.Set).List()
		groupIDs := make([]string, len(groups))
		for i, g := range groups {
			groupIDs[i] = g.(string)
		}
		schedule.TargetGroups = groupIDs
	}

	// Verificar se pelo menos um alvo foi especificado
	if len(schedule.TargetSystems) == 0 && len(schedule.TargetGroups) == 0 {
		return diag.FromErr(fmt.Errorf("pelo menos um sistema alvo ou grupo alvo deve ser especificado"))
	}

	// Serializar para JSON
	scheduleJSON, err := json.Marshal(schedule)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar agendamento de comando: %v", err))
	}

	// Atualizar agendamento via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando agendamento de comando: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/command/schedules/%s", id), scheduleJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar agendamento de comando: %v", err))
	}

	// Deserializar resposta
	var updatedSchedule CommandSchedule
	if err := json.Unmarshal(resp, &updatedSchedule); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceCommandScheduleRead(ctx, d, meta)
}

// resourceCommandScheduleDelete exclui um agendamento de comando
func resourceCommandScheduleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do agendamento de comando não fornecido"))
	}

	// Excluir agendamento via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo agendamento de comando: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/command/schedules/%s", id), nil)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Agendamento de comando %s não encontrado, considerando excluído", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir agendamento de comando: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
