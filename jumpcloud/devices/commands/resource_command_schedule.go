package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// ResourceCommandSchedule returns the resource for managing command schedules
func ResourceCommandSchedule() *schema.Resource {
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
				Description: "Name of the schedule",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the schedule",
			},
			"command_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the command to execute",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the schedule is active",
			},
			"schedule": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Cron expression or timestamp for scheduling",
			},
			"schedule_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"cron", "one-time"}, false),
				Description:  "Type of schedule: cron (recurring) or one-time",
			},
			"timezone": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "UTC",
				Description: "Timezone for the schedule (e.g., America/New_York)",
			},
			"target_systems": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IDs of systems where the command will be executed",
			},
			"target_groups": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IDs of groups where the command will be executed",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Schedule creation date",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of last schedule update",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manages command schedules in JumpCloud. This resource allows defining when and where commands will be executed.",
	}
}

// resourceCommandScheduleCreate creates a new command schedule
func resourceCommandScheduleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Build schedule
	schedule := &CommandSchedule{
		Name:         d.Get("name").(string),
		CommandID:    d.Get("command_id").(string),
		Enabled:      d.Get("enabled").(bool),
		Schedule:     d.Get("schedule").(string),
		ScheduleType: d.Get("schedule_type").(string),
		Timezone:     d.Get("timezone").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		schedule.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		schedule.OrgID = v.(string)
	}

	// Process target systems
	if v, ok := d.GetOk("target_systems"); ok {
		systems := v.(*schema.Set).List()
		systemIDs := make([]string, len(systems))
		for i, s := range systems {
			systemIDs[i] = s.(string)
		}
		schedule.TargetSystems = systemIDs
	}

	// Process target groups
	if v, ok := d.GetOk("target_groups"); ok {
		groups := v.(*schema.Set).List()
		groupIDs := make([]string, len(groups))
		for i, g := range groups {
			groupIDs[i] = g.(string)
		}
		schedule.TargetGroups = groupIDs
	}

	// Verify at least one target was specified
	if len(schedule.TargetSystems) == 0 && len(schedule.TargetGroups) == 0 {
		return diag.FromErr(fmt.Errorf("at least one target system or target group must be specified"))
	}

	// Serialize to JSON
	scheduleJSON, err := json.Marshal(schedule)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing command schedule: %v", err))
	}

	// Create schedule via API
	tflog.Debug(ctx, fmt.Sprintf("Creating command schedule: %s", schedule.Name))
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/command/schedules", scheduleJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating command schedule: %v", err))
	}

	// Deserialize response
	var createdSchedule CommandSchedule
	if err := json.Unmarshal(resp, &createdSchedule); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	if createdSchedule.ID == "" {
		return diag.FromErr(fmt.Errorf("command schedule created without an ID"))
	}

	d.SetId(createdSchedule.ID)
	return resourceCommandScheduleRead(ctx, d, meta)
}

// resourceCommandScheduleRead reads the details of a command schedule
func resourceCommandScheduleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get client
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("command schedule ID not provided"))
	}

	// Fetch schedule via API
	tflog.Debug(ctx, fmt.Sprintf("Reading command schedule with ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/command/schedules/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Command schedule %s not found, removing from state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading command schedule: %v", err))
	}

	// Deserialize response
	var schedule CommandSchedule
	if err := json.Unmarshal(resp, &schedule); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set values in state
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

// resourceCommandScheduleUpdate updates an existing command schedule
func resourceCommandScheduleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	scheduleID := d.Id()

	// Build updated schedule
	schedule := &CommandSchedule{
		ID:           scheduleID,
		Name:         d.Get("name").(string),
		CommandID:    d.Get("command_id").(string),
		Enabled:      d.Get("enabled").(bool),
		Schedule:     d.Get("schedule").(string),
		ScheduleType: d.Get("schedule_type").(string),
		Timezone:     d.Get("timezone").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		schedule.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		schedule.OrgID = v.(string)
	}

	// Process target systems
	if v, ok := d.GetOk("target_systems"); ok {
		systems := v.(*schema.Set).List()
		systemIDs := make([]string, len(systems))
		for i, s := range systems {
			systemIDs[i] = s.(string)
		}
		schedule.TargetSystems = systemIDs
	}

	// Process target groups
	if v, ok := d.GetOk("target_groups"); ok {
		groups := v.(*schema.Set).List()
		groupIDs := make([]string, len(groups))
		for i, g := range groups {
			groupIDs[i] = g.(string)
		}
		schedule.TargetGroups = groupIDs
	}

	// Verify at least one target was specified
	if len(schedule.TargetSystems) == 0 && len(schedule.TargetGroups) == 0 {
		return diag.FromErr(fmt.Errorf("at least one target system or target group must be specified"))
	}

	// Serialize to JSON
	scheduleJSON, err := json.Marshal(schedule)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing command schedule: %v", err))
	}

	// Update schedule via API
	tflog.Debug(ctx, fmt.Sprintf("Updating command schedule: %s", schedule.Name))
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/command/schedules/%s", scheduleID), scheduleJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating command schedule: %v", err))
	}

	return resourceCommandScheduleRead(ctx, d, meta)
}

// resourceCommandScheduleDelete deletes a command schedule
func resourceCommandScheduleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get client
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	scheduleID := d.Id()

	// Delete schedule via API
	tflog.Debug(ctx, fmt.Sprintf("Deleting command schedule: %s", scheduleID))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/command/schedules/%s", scheduleID), nil)
	if err != nil {
		// Check if the resource is already gone
		if common.IsNotFoundError(err) {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error deleting command schedule: %v", err))
	}

	// Clear ID from state
	d.SetId("")
	return diags
}
