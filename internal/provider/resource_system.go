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

func resourceSystem() *schema.Resource {
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
			},
			"agent_bound": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the system is bound to a JumpCloud agent",
			},
			"ssh_root_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether SSH root login is enabled for this system",
			},
			"organization_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The organization ID the system belongs to",
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceSystemCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	system := &System{
		DisplayName:                    d.Get("display_name").(string),
		AllowSshRootLogin:              d.Get("allow_ssh_root_login").(bool),
		AllowSshPasswordAuthentication: d.Get("allow_ssh_password_authentication").(bool),
		AllowMultiFactorAuthentication: d.Get("allow_multi_factor_authentication").(bool),
		Description:                    d.Get("description").(string),
	}

	if attr, ok := d.GetOk("attributes"); ok {
		system.Attributes = expandAttributes(attr.(map[string]interface{}))
	}

	if tags, ok := d.GetOk("tags"); ok {
		system.Tags = expandStringList(tags.([]interface{}))
	}

	tflog.Debug(ctx, "Creating JumpCloud system", map[string]interface{}{
		"display_name": system.DisplayName,
	})

	systemJSON, err := json.Marshal(system)
	if err != nil {
		return diag.Errorf("erro ao serializar sistema: %s", err)
	}

	// Fazer a requisição para criar o sistema
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/systems", systemJSON)
	if err != nil {
		return diag.Errorf("erro ao criar sistema: %s", err)
	}

	// Parse the response
	var createdSystem System
	err = json.Unmarshal(resp, &createdSystem)
	if err != nil {
		return diag.Errorf("erro ao deserializar resposta: %s", err)
	}

	// Set the ID from the response
	d.SetId(createdSystem.ID)

	return resourceSystemRead(ctx, d, m)
}

func resourceSystemRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	tflog.Debug(ctx, "Reading JumpCloud system", map[string]interface{}{
		"id": d.Id(),
	})

	// TODO: Implement the actual API call
	// This is a placeholder that would need to be replaced with a real API call
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/systems/%s", d.Id()), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading system: %v", err))
	}

	var system System
	if err := json.Unmarshal(resp, &system); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %v", err))
	}

	if err := d.Set("display_name", system.DisplayName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("system_type", system.SystemType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("os", system.OS); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("version", system.Version); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("agent_version", system.AgentVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("allow_ssh_root_login", system.AllowSshRootLogin); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("allow_ssh_password_authentication", system.AllowSshPasswordAuthentication); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("allow_multi_factor_authentication", system.AllowMultiFactorAuthentication); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", system.Tags); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", system.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("attributes", flattenAttributes(system.Attributes)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceSystemUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	system := &System{
		DisplayName:                    d.Get("display_name").(string),
		AllowSshRootLogin:              d.Get("allow_ssh_root_login").(bool),
		AllowSshPasswordAuthentication: d.Get("allow_ssh_password_authentication").(bool),
		AllowMultiFactorAuthentication: d.Get("allow_multi_factor_authentication").(bool),
		Description:                    d.Get("description").(string),
	}

	if attr, ok := d.GetOk("attributes"); ok {
		system.Attributes = expandAttributes(attr.(map[string]interface{}))
	}

	if tags, ok := d.GetOk("tags"); ok {
		system.Tags = expandStringList(tags.([]interface{}))
	}

	tflog.Debug(ctx, "Updating JumpCloud system", map[string]interface{}{
		"id": d.Id(),
	})

	// TODO: Implement the actual API call
	// This is a placeholder that would need to be replaced with a real API call
	_, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/systems/%s", d.Id()), system)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating system: %v", err))
	}

	return resourceSystemRead(ctx, d, m)
}

func resourceSystemDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	tflog.Debug(ctx, "Deleting JumpCloud system", map[string]interface{}{
		"id": d.Id(),
	})

	// TODO: Implement the actual API call
	// This is a placeholder that would need to be replaced with a real API call
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/systems/%s", d.Id()), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting system: %v", err))
	}

	d.SetId("")

	return nil
}

// expandStringList converts a list of interfaces to a list of strings
func expandStringList(list []interface{}) []string {
	result := make([]string, len(list))
	for i, v := range list {
		result[i] = v.(string)
	}
	return result
}

// flattenStringList converts a list of strings to a list of interfaces
func flattenStringList(list []string) []interface{} {
	result := make([]interface{}, len(list))
	for i, v := range list {
		result[i] = v
	}
	return result
}
