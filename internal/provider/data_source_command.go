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
)

func dataSourceCommand() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCommandRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"name"},
				Description:   "ID do comando",
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"id"},
				Description:   "Nome do comando",
			},
			"command": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "O comando a ser executado",
			},
			"command_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Tipo do comando (linux, windows, mac)",
			},
			"user": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Usuário que executará o comando",
			},
			"schedule": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Agendamento do comando",
			},
			"schedule_repeat": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Tipo de repetição do agendamento",
			},
			"trigger": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Gatilho de execução do comando",
			},
			"shell": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Shell usado para executar o comando",
			},
			"sudo": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Se o comando deve ser executado com privilégios sudo",
			},
			"launch_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Tipo de lançamento do comando",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Tempo limite de execução do comando em segundos",
			},
			"files": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Lista de arquivos associados ao comando",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"environments": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Lista de ambientes para execução do comando",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Descrição do comando",
			},
			"attributes": {
				Type:        schema.TypeMap,
				Computed:    true,
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
			"target_systems": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Sistemas associados ao comando",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"target_groups": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Grupos de sistemas associados ao comando",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
		Description: "Use este data source para buscar informações sobre um comando existente no JumpCloud.",
	}
}

func dataSourceCommandRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Lendo data source de comando do JumpCloud")

	var diags diag.Diagnostics

	c, convDiags := ConvertToClientInterface(meta)
	if convDiags != nil {
		return convDiags
	}

	var commandID string
	var resp []byte
	var err error

	// Buscar por ID ou por nome
	if id, ok := d.GetOk("id"); ok {
		commandID = id.(string)
		resp, err = c.DoRequest(http.MethodGet, fmt.Sprintf("/api/commands/%s", commandID), nil)
	} else if name, ok := d.GetOk("name"); ok {
		// Buscar comando por nome: primeiro obtemos todos os comandos e filtramos pelo nome
		resp, err = c.DoRequest(http.MethodGet, "/api/commands", nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao buscar comandos: %v", err))
		}

		// Decodificar a resposta como uma lista de comandos
		var commands []Command
		if err := json.Unmarshal(resp, &commands); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
		}

		// Procurar comando pelo nome
		commandName := name.(string)
		for _, cmd := range commands {
			if cmd.Name == commandName {
				commandID = cmd.ID
				// Agora que temos o ID, buscamos os detalhes específicos do comando
				resp, err = c.DoRequest(http.MethodGet, fmt.Sprintf("/api/commands/%s", commandID), nil)
				break
			}
		}

		if commandID == "" {
			return diag.FromErr(fmt.Errorf("comando com nome '%s' não encontrado", commandName))
		}
	} else {
		return diag.FromErr(fmt.Errorf("deve ser fornecido um ID ou um nome para buscar um comando"))
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao buscar comando: %v", err))
	}

	// Decodificar a resposta
	var command Command
	if err := json.Unmarshal(resp, &command); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir ID do recurso
	d.SetId(command.ID)

	// Definir atributos no estado
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

	for k, v := range fields {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.FromErr(fmt.Errorf("erro ao definir campo %s: %v", k, err))...)
		}
	}

	// Buscar metadados adicionais como created
	metaResp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/commands/%s/metadata", commandID), nil)
	if err == nil {
		var metadata struct {
			Created time.Time `json:"created"`
		}
		if err := json.Unmarshal(metaResp, &metadata); err == nil {
			d.Set("created", metadata.Created.Format(time.RFC3339))
		}
	}

	// Buscar informações sobre sistemas e grupos associados
	assocResp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/commands/%s/associations", commandID), nil)
	if err == nil {
		var associations struct {
			Results []struct {
				To struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"to"`
			} `json:"results"`
		}
		if err := json.Unmarshal(assocResp, &associations); err == nil {
			var systems []string
			var groups []string

			for _, assoc := range associations.Results {
				if assoc.To.Type == "system" {
					systems = append(systems, assoc.To.ID)
				} else if assoc.To.Type == "system_group" {
					groups = append(groups, assoc.To.ID)
				}
			}

			d.Set("target_systems", systems)
			d.Set("target_groups", groups)
		}
	}

	return diags
}
