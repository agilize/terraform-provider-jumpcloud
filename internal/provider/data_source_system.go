package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSystem() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSystemRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"system_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"system_id", "display_name"},
			},
			"display_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"system_id", "display_name"},
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
				Computed: true,
			},
			"allow_ssh_password_authentication": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"allow_multi_factor_authentication": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tags": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attributes": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"agent_bound": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ssh_root_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Computed: true,
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
	}
}

func dataSourceSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	var path string

	if systemID, ok := d.GetOk("system_id"); ok {
		path = fmt.Sprintf("/api/systems/%s", systemID.(string))
	} else if displayName, ok := d.GetOk("display_name"); ok {
		// In real implementation, we would need to search for systems by display name
		// This is a placeholder that would need to be replaced with a real API call
		path = fmt.Sprintf("/api/search/systems?displayName=%s", displayName.(string))
	} else {
		return diag.FromErr(fmt.Errorf("one of system_id or display_name must be provided"))
	}

	tflog.Debug(ctx, "Reading JumpCloud system data source", map[string]interface{}{
		"path": path,
	})

	// Usar a interface DoRequest em vez de acessar o cliente diretamente
	resp, err := c.DoRequest(http.MethodGet, path, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading system: %v", err))
	}

	var system System
	if err := json.Unmarshal(resp, &system); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %v", err))
	}

	d.SetId(system.ID)

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
