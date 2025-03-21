package systems

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// System represents a JumpCloud system (device)
type System struct {
	ID                             string                 `json:"_id,omitempty"`
	DisplayName                    string                 `json:"displayName,omitempty"`
	SystemType                     string                 `json:"systemType,omitempty"`
	OS                             string                 `json:"os,omitempty"`
	Version                        string                 `json:"version,omitempty"`
	AgentVersion                   string                 `json:"agentVersion,omitempty"`
	AllowSshRootLogin              bool                   `json:"allowSshRootLogin,omitempty"`
	AllowSshPasswordAuthentication bool                   `json:"allowSshPasswordAuthentication,omitempty"`
	AllowMultiFactorAuthentication bool                   `json:"allowMultiFactorAuthentication,omitempty"`
	Tags                           []string               `json:"tags,omitempty"`
	Description                    string                 `json:"description,omitempty"`
	Attributes                     map[string]interface{} `json:"attributes,omitempty"`
}

func ResourceSystem() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSystemCreate,
		ReadContext:   resourceSystemRead,
		UpdateContext: resourceSystemUpdate,
		DeleteContext: resourceSystemDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"system_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"os": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"agent_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"allow_ssh_root_login": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"allow_ssh_password_authentication": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"allow_multi_factor_authentication": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Custom attributes for the system (key-value pairs)",
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_contact": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"remote_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"has_active_agent": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"mdm_managed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enrollment_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"serial_number": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSystemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Build system object from resource data
	system := &System{
		DisplayName:                    d.Get("display_name").(string),
		AllowSshRootLogin:              d.Get("allow_ssh_root_login").(bool),
		AllowSshPasswordAuthentication: d.Get("allow_ssh_password_authentication").(bool),
		AllowMultiFactorAuthentication: d.Get("allow_multi_factor_authentication").(bool),
		Description:                    d.Get("description").(string),
	}

	// Process tags
	if tagsRaw, ok := d.GetOk("tags"); ok {
		tags := expandStringList(tagsRaw.([]interface{}))
		system.Tags = tags
	}

	// Process attributes
	if attrRaw, ok := d.GetOk("attributes"); ok {
		attributes := make(map[string]interface{})
		for k, v := range attrRaw.(map[string]interface{}) {
			attributes[k] = v
		}
		system.Attributes = attributes
	}

	// Convert system to JSON
	systemJSON, err := json.Marshal(system)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing system: %v", err))
	}

	// Create system via API
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/systems", systemJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating system: %v", err))
	}

	// Decode response
	var newSystem System
	if err := json.Unmarshal(resp, &newSystem); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing system response: %v", err))
	}

	// Set ID in resource data
	d.SetId(newSystem.ID)

	// Read the system to set all the computed fields
	return resourceSystemRead(ctx, d, meta)
}

func resourceSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	systemID := d.Id()

	// Get system via API
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/systems/%s", systemID), nil)
	if err != nil {
		// Handle 404 specifically to mark the resource as removed
		if err.Error() == "status code 404" {
			tflog.Warn(ctx, fmt.Sprintf("System %s not found, removing from state", systemID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading system %s: %v", systemID, err))
	}

	// Decode response
	var system System
	if err := json.Unmarshal(resp, &system); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing system response: %v", err))
	}

	// Set fields in resource data
	d.Set("display_name", system.DisplayName)
	d.Set("system_type", system.SystemType)
	d.Set("os", system.OS)
	d.Set("version", system.Version)
	d.Set("agent_version", system.AgentVersion)
	d.Set("allow_ssh_root_login", system.AllowSshRootLogin)
	d.Set("allow_ssh_password_authentication", system.AllowSshPasswordAuthentication)
	d.Set("allow_multi_factor_authentication", system.AllowMultiFactorAuthentication)
	d.Set("description", system.Description)

	// Handle tags
	if system.Tags != nil {
		d.Set("tags", flattenStringList(system.Tags))
	}

	// Handle attributes
	if system.Attributes != nil {
		attributes := make(map[string]interface{})
		for k, v := range system.Attributes {
			attributes[k] = fmt.Sprintf("%v", v)
		}
		d.Set("attributes", attributes)
	}

	return diags
}

func resourceSystemUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	systemID := d.Id()

	// Build system object from resource data
	system := &System{
		DisplayName:                    d.Get("display_name").(string),
		AllowSshRootLogin:              d.Get("allow_ssh_root_login").(bool),
		AllowSshPasswordAuthentication: d.Get("allow_ssh_password_authentication").(bool),
		AllowMultiFactorAuthentication: d.Get("allow_multi_factor_authentication").(bool),
		Description:                    d.Get("description").(string),
	}

	// Process tags
	if tagsRaw, ok := d.GetOk("tags"); ok {
		tags := expandStringList(tagsRaw.([]interface{}))
		system.Tags = tags
	}

	// Process attributes
	if attrRaw, ok := d.GetOk("attributes"); ok {
		attributes := make(map[string]interface{})
		for k, v := range attrRaw.(map[string]interface{}) {
			attributes[k] = v
		}
		system.Attributes = attributes
	}

	// Convert system to JSON
	systemJSON, err := json.Marshal(system)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing system: %v", err))
	}

	// Update system via API
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/systems/%s", systemID), systemJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating system %s: %v", systemID, err))
	}

	return resourceSystemRead(ctx, d, meta)
}

func resourceSystemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	systemID := d.Id()

	// Delete system via API
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/systems/%s", systemID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting system %s: %v", systemID, err))
	}

	// Set ID to empty to signify resource has been removed
	d.SetId("")

	return diags
}

func expandStringList(list []interface{}) []string {
	result := make([]string, 0, len(list))
	for _, item := range list {
		result = append(result, item.(string))
	}
	return result
}

func flattenStringList(list []string) []interface{} {
	result := make([]interface{}, 0, len(list))
	for _, item := range list {
		result = append(result, item)
	}
	return result
}
