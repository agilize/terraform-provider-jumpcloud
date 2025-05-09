package devices

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

// DataSourceSystem returns the schema for the JumpCloud System data source
func DataSourceSystem() *schema.Resource {
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
				Description: "Custom attributes for the system (key-value pairs)",
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

func dataSourceSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	var path string

	if systemID, ok := d.GetOk("system_id"); ok {
		path = fmt.Sprintf("/api/v2/systems/%s", systemID.(string))
	} else if displayName, ok := d.GetOk("display_name"); ok {
		// Search for systems by display name
		path = fmt.Sprintf("/api/v2/search/systems?filter=displayName:%s", displayName.(string))
	} else {
		return diag.FromErr(fmt.Errorf("one of system_id or display_name must be provided"))
	}

	tflog.Debug(ctx, "Reading JumpCloud system data source", map[string]interface{}{
		"path": path,
	})

	resp, err := c.DoRequest(http.MethodGet, path, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading system: %v", err))
	}

	var system common.System
	if displayName, ok := d.GetOk("display_name"); ok {
		// When searching by display name, we need to handle the search results format
		var searchResults struct {
			TotalCount int             `json:"totalCount"`
			Results    []common.System `json:"results"`
		}

		if err := json.Unmarshal(resp, &searchResults); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing search response: %v", err))
		}

		if searchResults.TotalCount == 0 {
			return diag.FromErr(fmt.Errorf("no system found with display name: %s", displayName))
		}

		if searchResults.TotalCount > 1 {
			tflog.Warn(ctx, fmt.Sprintf("Found multiple systems with display name %s, using the first one", displayName))
		}

		system = searchResults.Results[0]
	} else {
		// Direct system lookup by ID
		if err := json.Unmarshal(resp, &system); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing response: %v", err))
		}
	}

	d.SetId(system.ID)

	// Set fields in resource data
	if err := d.Set("display_name", system.DisplayName); err != nil {
		return diag.FromErr(fmt.Errorf("error setting display_name: %v", err))
	}

	if err := d.Set("system_type", system.SystemType); err != nil {
		return diag.FromErr(fmt.Errorf("error setting system_type: %v", err))
	}

	if err := d.Set("os", system.OS); err != nil {
		return diag.FromErr(fmt.Errorf("error setting os: %v", err))
	}

	if err := d.Set("version", system.Version); err != nil {
		return diag.FromErr(fmt.Errorf("error setting version: %v", err))
	}

	if err := d.Set("agent_version", system.AgentVersion); err != nil {
		return diag.FromErr(fmt.Errorf("error setting agent_version: %v", err))
	}

	if err := d.Set("allow_ssh_root_login", system.AllowSshRootLogin); err != nil {
		return diag.FromErr(fmt.Errorf("error setting allow_ssh_root_login: %v", err))
	}

	if err := d.Set("allow_ssh_password_authentication", system.AllowSshPasswordAuthentication); err != nil {
		return diag.FromErr(fmt.Errorf("error setting allow_ssh_password_authentication: %v", err))
	}

	if err := d.Set("allow_multi_factor_authentication", system.AllowMultiFactorAuthentication); err != nil {
		return diag.FromErr(fmt.Errorf("error setting allow_multi_factor_authentication: %v", err))
	}

	if err := d.Set("description", system.Description); err != nil {
		return diag.FromErr(fmt.Errorf("error setting description: %v", err))
	}

	// Set tags if they exist
	if len(system.Tags) > 0 {
		if err := d.Set("tags", common.FlattenStringList(system.Tags)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting tags: %v", err))
		}
	}

	// Handle attributes
	if system.Attributes != nil {
		attributes := make(map[string]interface{})
		for k, v := range system.Attributes {
			attributes[k] = fmt.Sprintf("%v", v)
		}
		if err := d.Set("attributes", attributes); err != nil {
			return diag.FromErr(fmt.Errorf("error setting attributes: %v", err))
		}
	}

	return diags
}
