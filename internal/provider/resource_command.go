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

// Command representa um comando no JumpCloud
type Command struct {
	ID             string                 `json:"_id,omitempty"`
	Name           string                 `json:"name"`
	Command        string                 `json:"command"`
	CommandType    string                 `json:"commandType"`
	User           string                 `json:"user,omitempty"`
	Schedule       string                 `json:"schedule,omitempty"`
	ScheduleRepeat string                 `json:"scheduleRepeatType,omitempty"`
	Trigger        string                 `json:"trigger,omitempty"`
	Shell          string                 `json:"shell,omitempty"`
	Sudo           bool                   `json:"sudo,omitempty"`
	LaunchType     string                 `json:"launchType,omitempty"`
	Timeout        int                    `json:"timeout,omitempty"`
	Files          []string               `json:"files,omitempty"`
	Environments   []string               `json:"environments,omitempty"`
	Description    string                 `json:"description,omitempty"`
	Attributes     map[string]interface{} `json:"attributes,omitempty"`
}

// resourceCommand retorna o resource para gerenciar comandos
func resourceCommand() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCommandCreate,
		ReadContext:   resourceCommandRead,
		UpdateContext: resourceCommandUpdate,
		DeleteContext: resourceCommandDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome do comando",
			},
			"command": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "O comando a ser executado",
			},
			"command_type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Tipo do comando (linux, windows, mac)",
				ValidateFunc: validation.StringInSlice([]string{"linux", "windows", "mac"}, false),
			},
			"user": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "root",
				Description: "Usuário que executará o comando (padrão: root)",
			},
			"schedule": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Agendamento do comando no formato cron",
			},
			"schedule_repeat": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Tipo de repetição do agendamento",
				ValidateFunc: validation.StringInSlice([]string{"once", "daily", "weekly", "monthly"}, false),
			},
			"trigger": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "manual",
				Description:  "Gatilho de execução do comando",
				ValidateFunc: validation.StringInSlice([]string{"manual", "automatic", "deadline", "periodic"}, false),
			},
			"shell": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Shell usado para executar o comando",
				ValidateFunc: validation.StringInSlice([]string{"bash", "powershell", "sh", "zsh"}, false),
			},
			"sudo": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Se o comando deve ser executado com privilégios sudo",
			},
			"launch_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "manual",
				Description:  "Tipo de lançamento do comando",
				ValidateFunc: validation.StringInSlice([]string{"manual", "auto"}, false),
			},
			"timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      120,
				Description:  "Tempo limite de execução do comando em segundos",
				ValidateFunc: validation.IntBetween(30, 3600),
			},
			"files": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Lista de arquivos associados ao comando",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"environments": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Lista de ambientes para execução do comando",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição do comando",
			},
			"attributes": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Atributos personalizados do comando",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação do comando",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Gerencia comandos no JumpCloud. Este recurso permite criar, atualizar e excluir comandos para execução em sistemas.",
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Update: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

// resourceCommandCreate cria um novo comando no JumpCloud
func resourceCommandCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Criando comando no JumpCloud")

	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	// Criar objeto Command a partir dos dados do resource
	cmd := &Command{
		Name:           d.Get("name").(string),
		Command:        d.Get("command").(string),
		CommandType:    d.Get("command_type").(string),
		User:           d.Get("user").(string),
		Schedule:       d.Get("schedule").(string),
		ScheduleRepeat: d.Get("schedule_repeat").(string),
		Trigger:        d.Get("trigger").(string),
		Shell:          d.Get("shell").(string),
		Sudo:           d.Get("sudo").(bool),
		LaunchType:     d.Get("launch_type").(string),
		Timeout:        d.Get("timeout").(int),
		Description:    d.Get("description").(string),
	}

	// Processar listas
	if v, ok := d.GetOk("files"); ok {
		cmd.Files = expandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("environments"); ok {
		cmd.Environments = expandStringList(v.([]interface{}))
	}

	// Processar atributos personalizados, se houver
	if v, ok := d.GetOk("attributes"); ok {
		cmd.Attributes = expandAttributes(v.(map[string]interface{}))
	}

	// Converter para JSON
	jsonData, err := json.Marshal(cmd)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar comando: %v", err))
	}

	// Enviar requisição para criar o comando
	resp, err := c.DoRequest(http.MethodPost, "/api/commands", jsonData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar comando: %v", err))
	}

	// Deserializar a resposta
	var createdCommand Command
	if err := json.Unmarshal(resp, &createdCommand); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir ID do recurso
	d.SetId(createdCommand.ID)

	// Ler o recurso para atualizar o estado
	return resourceCommandRead(ctx, d, meta)
}

// resourceCommandRead lê as informações de um comando do JumpCloud
func resourceCommandRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Lendo comando do JumpCloud")

	var diags diag.Diagnostics

	c, convDiags := ConvertToClientInterface(meta)
	if convDiags != nil {
		return convDiags
	}

	// Buscar informações do comando pelo ID
	commandID := d.Id()
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/commands/%s", commandID), nil)
	if err != nil {
		// Verificar se o comando não existe mais
		if isNotFoundError(err) {
			tflog.Warn(ctx, "Comando não encontrado, removendo do estado", map[string]interface{}{
				"id": commandID,
			})
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao buscar comando: %v", err))
	}

	// Deserializar a resposta
	var command Command
	if err := json.Unmarshal(resp, &command); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Atualizar o estado do recurso
	fields := map[string]interface{}{
		"name":            command.Name,
		"command":         command.Command,
		"command_type":    command.CommandType,
		"user":            command.User,
		"schedule":        command.Schedule,
		"schedule_repeat": command.ScheduleRepeat,
		"trigger":         command.Trigger,
		"shell":           command.Shell,
		"sudo":            command.Sudo,
		"launch_type":     command.LaunchType,
		"timeout":         command.Timeout,
		"description":     command.Description,
		"files":           flattenStringList(command.Files),
		"environments":    flattenStringList(command.Environments),
		"attributes":      flattenAttributes(command.Attributes),
	}

	// Definir campos no estado
	for k, v := range fields {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.FromErr(fmt.Errorf("erro ao definir campo %s: %v", k, err))...)
		}
	}

	// Obter metadados adicionais
	metaResp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/commands/%s/metadata", commandID), nil)
	if err == nil {
		var metadata struct {
			Created time.Time `json:"created"`
		}
		if err := json.Unmarshal(metaResp, &metadata); err == nil {
			d.Set("created", metadata.Created.Format(time.RFC3339))
		}
	}

	return diags
}

// resourceCommandUpdate atualiza um comando existente no JumpCloud
func resourceCommandUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Atualizando comando no JumpCloud")

	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	// Verificar se houve mudanças nos campos
	if !d.HasChanges("name", "command", "command_type", "user", "schedule", "schedule_repeat",
		"trigger", "shell", "sudo", "launch_type", "timeout", "files", "environments",
		"description", "attributes") {
		return resourceCommandRead(ctx, d, meta)
	}

	// Preparar objeto de atualização
	cmd := &Command{
		Name:           d.Get("name").(string),
		Command:        d.Get("command").(string),
		CommandType:    d.Get("command_type").(string),
		User:           d.Get("user").(string),
		Schedule:       d.Get("schedule").(string),
		ScheduleRepeat: d.Get("schedule_repeat").(string),
		Trigger:        d.Get("trigger").(string),
		Shell:          d.Get("shell").(string),
		Sudo:           d.Get("sudo").(bool),
		LaunchType:     d.Get("launch_type").(string),
		Timeout:        d.Get("timeout").(int),
		Description:    d.Get("description").(string),
	}

	// Processar listas
	if v, ok := d.GetOk("files"); ok {
		cmd.Files = expandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("environments"); ok {
		cmd.Environments = expandStringList(v.([]interface{}))
	}

	// Processar atributos personalizados, se houver
	if v, ok := d.GetOk("attributes"); ok {
		cmd.Attributes = expandAttributes(v.(map[string]interface{}))
	}

	// Converter para JSON
	jsonData, err := json.Marshal(cmd)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar comando: %v", err))
	}

	// Enviar requisição de atualização
	commandID := d.Id()
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/commands/%s", commandID), jsonData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar comando: %v", err))
	}

	// Ler o recurso para atualizar o estado
	return resourceCommandRead(ctx, d, meta)
}

// resourceCommandDelete exclui um comando do JumpCloud
func resourceCommandDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Excluindo comando do JumpCloud")

	var diags diag.Diagnostics

	c, convDiags := ConvertToClientInterface(meta)
	if convDiags != nil {
		return convDiags
	}

	// Enviar requisição para excluir o comando
	commandID := d.Id()
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/commands/%s", commandID), nil)
	if err != nil {
		// Se o recurso já foi excluído, não considerar como erro
		if isNotFoundError(err) {
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir comando: %v", err))
	}

	// Limpar o ID para indicar que o recurso foi excluído
	d.SetId("")

	return diags
}
