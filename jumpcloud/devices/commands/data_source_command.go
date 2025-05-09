package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// DataSourceCommand returns the schema for the JumpCloud command data source
func DataSourceCommand() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCommandRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"name"},
				Description:   "ID of the command",
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"id"},
				Description:   "Name of the command",
			},
			"command": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The command to be executed",
			},
			"command_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of command (linux, windows, mac)",
			},
			"user": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "User that will execute the command",
			},
			"schedule": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Command schedule",
			},
			"schedule_repeat": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Schedule repeat type",
			},
			"trigger": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Command execution trigger",
			},
			"shell": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Shell used to execute the command",
			},
			"sudo": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the command should be executed with sudo privileges",
			},
			"launch_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Command launch type",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Command execution timeout in seconds",
			},
			"files": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of files associated with the command",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"environments": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of environments for command execution",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Command description",
			},
			"attributes": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Custom command attributes",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Command creation date",
			},
			"target_systems": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Systems associated with the command",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"target_groups": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "System groups associated with the command",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
		Description: "Use this data source to retrieve information about an existing command in JumpCloud.",
	}
}

func dataSourceCommandRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Reading command data source from JumpCloud")

	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	var commandID string
	var resp []byte
	var err error

	// Search by ID or by name
	if id, ok := d.GetOk("id"); ok {
		commandID = id.(string)
		resp, err = c.DoRequest(http.MethodGet, fmt.Sprintf("/api/commands/%s", commandID), nil)
	} else if name, ok := d.GetOk("name"); ok {
		// Search command by name: first get all commands and filter by name
		resp, err = c.DoRequest(http.MethodGet, "/api/commands", nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error fetching commands: %v", err))
		}

		// Decode the response as a list of commands
		var commands []Command
		if err := json.Unmarshal(resp, &commands); err != nil {
			return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
		}

		// Find command by name
		commandName := name.(string)
		for _, cmd := range commands {
			if cmd.Name == commandName {
				commandID = cmd.ID
				// Now that we have the ID, get the specific command details
				resp, err = c.DoRequest(http.MethodGet, fmt.Sprintf("/api/commands/%s", commandID), nil)
				break
			}
		}

		if commandID == "" {
			return diag.FromErr(fmt.Errorf("command with name '%s' not found", commandName))
		}
	} else {
		return diag.FromErr(fmt.Errorf("either id or name must be provided to look up a command"))
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching command: %v", err))
	}

	// Decode the response
	var command Command
	if err := json.Unmarshal(resp, &command); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set resource ID
	d.SetId(command.ID)

	// Set attributes in state
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
		"files":           command.Files,
		"environments":    command.Environments,
	}

	for k, v := range fields {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.FromErr(fmt.Errorf("error setting field %s: %v", k, err))...)
		}
	}

	// Handle attributes specially
	if command.Attributes != nil {
		attributes := common.FlattenAttributes(command.Attributes)
		if err := d.Set("attributes", attributes); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	// Fetch additional metadata such as created date
	metaResp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/commands/%s/metadata", commandID), nil)
	if err == nil {
		var metadata struct {
			Created time.Time `json:"created"`
		}
		if err := json.Unmarshal(metaResp, &metadata); err == nil {
			if err := d.Set("created", metadata.Created.Format(time.RFC3339)); err != nil {
				return diag.FromErr(fmt.Errorf("error setting created: %v", err))
			}
		}
	}

	// Fetch information about associated systems and groups
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
				switch assoc.To.Type {
				case "system":
					systems = append(systems, assoc.To.ID)
				case "system_group":
					groups = append(groups, assoc.To.ID)
				}
			}

			if len(systems) > 0 {
				if err := d.Set("target_systems", systems); err != nil {
					diags = append(diags, diag.FromErr(err)...)
				}
			}

			if len(groups) > 0 {
				if err := d.Set("target_groups", groups); err != nil {
					diags = append(diags, diag.FromErr(err)...)
				}
			}
		}
	}

	return diags
}
