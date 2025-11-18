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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// ResourceCommand returns the resource schema for JumpCloud commands
func ResourceCommand() *schema.Resource {
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
				Description: "Name of the command",
			},
			"command": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The command to be executed",
			},
			"command_type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Type of command (linux, windows, mac)",
				ValidateFunc: validation.StringInSlice([]string{"linux", "windows", "mac"}, false),
			},
			"user": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "root",
				Description: "User that will execute the command (default: root)",
			},
			"schedule": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Command schedule in cron format",
			},
			"schedule_repeat": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Schedule repeat type",
				ValidateFunc: validation.StringInSlice([]string{"once", "daily", "weekly", "monthly"}, false),
			},
			"trigger": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "manual",
				Description:  "Command execution trigger",
				ValidateFunc: validation.StringInSlice([]string{"manual", "automatic", "deadline", "periodic"}, false),
			},
			"shell": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Shell used to execute the command",
				ValidateFunc: validation.StringInSlice([]string{"bash", "powershell", "sh", "zsh"}, false),
			},
			"sudo": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the command should be executed with sudo privileges",
			},
			"launch_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "manual",
				Description:  "Command launch type",
				ValidateFunc: validation.StringInSlice([]string{"manual", "auto"}, false),
			},
			"timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      120,
				Description:  "Command execution timeout in seconds",
				ValidateFunc: validation.IntBetween(30, 3600),
			},
			"files": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of files associated with the command",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"environments": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of environments for command execution",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Command description",
			},
			"attributes": {
				Type:        schema.TypeMap,
				Optional:    true,
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
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manages commands in JumpCloud. This resource allows creating, updating, and deleting commands for execution on systems.",
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Update: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

// resourceCommandCreate creates a new command in JumpCloud
func resourceCommandCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Creating command in JumpCloud")

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Create Command object from resource data
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

	// Process lists
	if v, ok := d.GetOk("files"); ok {
		cmd.Files = common.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("environments"); ok {
		cmd.Environments = common.ExpandStringList(v.([]interface{}))
	}

	// Process custom attributes, if any
	if v, ok := d.GetOk("attributes"); ok {
		cmd.Attributes = common.ExpandAttributes(v.(map[string]interface{}))
	}

	// Convert to JSON
	jsonData, err := json.Marshal(cmd)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing command: %v", err))
	}

	// Send request to create the command
	resp, err := c.DoRequest(http.MethodPost, "/api/commands", jsonData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating command: %v", err))
	}

	// Deserialize the response
	var createdCommand Command
	if err := json.Unmarshal(resp, &createdCommand); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set resource ID
	d.SetId(createdCommand.ID)

	// Read the resource to update the state
	return resourceCommandRead(ctx, d, meta)
}

// resourceCommandRead reads command information from JumpCloud
func resourceCommandRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Reading command from JumpCloud")

	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Fetch command information by ID
	commandID := d.Id()
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/commands/%s", commandID), nil)
	if err != nil {
		// Check if the command no longer exists
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, "Command not found, removing from state", map[string]interface{}{
				"id": commandID,
			})
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error fetching command: %v", err))
	}

	// Deserialize the response
	var command Command
	if err := json.Unmarshal(resp, &command); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set fields in state
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
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	// Handle attributes specially
	if command.Attributes != nil {
		attributes := common.FlattenAttributes(command.Attributes)
		if err := d.Set("attributes", attributes); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	// Fetch metadata to get creation date
	metaResp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/commands/%s/metadata", commandID), nil)
	if err == nil {
		var metadata struct {
			Created time.Time `json:"created"`
		}
		if err := json.Unmarshal(metaResp, &metadata); err == nil {
			if err := d.Set("created", metadata.Created.Format(time.RFC3339)); err != nil {
				diags = append(diags, diag.FromErr(err)...)
			}
		}
	}

	return diags
}

// resourceCommandUpdate updates an existing command in JumpCloud
func resourceCommandUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Updating command in JumpCloud")

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	commandID := d.Id()

	// Create Command object from resource data
	cmd := &Command{
		ID:             commandID,
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

	// Process lists
	if v, ok := d.GetOk("files"); ok {
		cmd.Files = common.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("environments"); ok {
		cmd.Environments = common.ExpandStringList(v.([]interface{}))
	}

	// Process custom attributes, if any
	if v, ok := d.GetOk("attributes"); ok {
		cmd.Attributes = common.ExpandAttributes(v.(map[string]interface{}))
	}

	// Convert to JSON
	jsonData, err := json.Marshal(cmd)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing command: %v", err))
	}

	// Send request to update the command
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/commands/%s", commandID), jsonData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating command: %v", err))
	}

	return resourceCommandRead(ctx, d, meta)
}

// resourceCommandDelete deletes a command from JumpCloud
func resourceCommandDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Deleting command from JumpCloud")

	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	commandID := d.Id()

	// Send request to delete the command
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/commands/%s", commandID), nil)
	if err != nil {
		// If the resource is already gone, don't return an error
		if common.IsNotFoundError(err) {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error deleting command: %v", err))
	}

	// Clear ID to mark resource as deleted
	d.SetId("")

	return diags
}
