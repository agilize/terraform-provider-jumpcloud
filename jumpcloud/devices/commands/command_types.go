package commands

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// CommandTypes returns a map of valid command types
func CommandTypes() map[string]bool {
	return map[string]bool{
		"windows": true,
		"linux":   true,
		"mac":     true,
	}
}

// LaunchTypes returns a map of valid launch types
func LaunchTypes() map[string]bool {
	return map[string]bool{
		"manual":   true,
		"trigger":  true,
		"schedule": true,
		"repeated": true,
	}
}

// TriggerTypes returns a map of valid trigger types
func TriggerTypes() map[string]bool {
	return map[string]bool{
		"date":           true,
		"time":           true,
		"interval":       true,
		"session_start":  true,
		"session_stop":   true,
		"network_join":   true,
		"network_leave":  true,
		"filesystem":     true,
		"pending_reboot": true,
	}
}

// CommonCommandSchema returns the schema fields common to all command-related resources
func CommonCommandSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"command": {
			Type:     schema.TypeString,
			Required: true,
		},
		"command_type": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice(mapKeysToSlice(CommandTypes()), false),
		},
		"user": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "root",
		},
		"sudo": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"shell": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"timeout": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "0",
			ValidateFunc: validation.StringIsNotEmpty,
		},
		"launch_type": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "manual",
			ValidateFunc: validation.StringInSlice(mapKeysToSlice(LaunchTypes()), false),
		},
		"trigger": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice(mapKeysToSlice(TriggerTypes()), false),
		},
		"schedule": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"files": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"organization_id": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"template_variables": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"created": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"updated": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

// Command represents a JumpCloud command
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
	Organization   string                 `json:"organization,omitempty"`
	Created        string                 `json:"created,omitempty"`
	Updated        string                 `json:"updated,omitempty"`
}

// CommandSchedule represents a JumpCloud command schedule
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

// Helper function to convert a map[string]bool to []string
func mapKeysToSlice(m map[string]bool) []string {
	var slice []string
	for k := range m {
		slice = append(slice, k)
	}
	return slice
}
