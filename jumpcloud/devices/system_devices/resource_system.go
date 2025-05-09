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
	c, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Build system object from resource data
	system := &common.System{
		DisplayName:                    d.Get("display_name").(string),
		AllowSshRootLogin:              d.Get("allow_ssh_root_login").(bool),
		AllowSshPasswordAuthentication: d.Get("allow_ssh_password_authentication").(bool),
		AllowMultiFactorAuthentication: d.Get("allow_multi_factor_authentication").(bool),
		Description:                    d.Get("description").(string),
	}

	// Handle tags
	if tagsRaw, ok := d.GetOk("tags"); ok {
		system.Tags = common.ExpandStringList(tagsRaw.([]interface{}))
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
	var newSystem common.System
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

	c, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do sistema não fornecido"))
	}

	// Buscar sistema via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo sistema com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/systems/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Sistema %s não encontrado, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler sistema: %v", err))
	}

	// Deserializar resposta
	var system common.System
	if err := json.Unmarshal(resp, &system); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Set tags if they exist
	if len(system.Tags) > 0 {
		if err := d.Set("tags", common.FlattenStringList(system.Tags)); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao definir tags: %v", err))
		}
	}

	// Set fields in resource data
	if err := d.Set("display_name", system.DisplayName); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir display_name: %v", err))
	}

	if err := d.Set("system_type", system.SystemType); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir system_type: %v", err))
	}

	if err := d.Set("os", system.OS); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir os: %v", err))
	}

	if err := d.Set("version", system.Version); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir version: %v", err))
	}

	if err := d.Set("agent_version", system.AgentVersion); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir agent_version: %v", err))
	}

	if err := d.Set("allow_ssh_root_login", system.AllowSshRootLogin); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir allow_ssh_root_login: %v", err))
	}

	if err := d.Set("allow_ssh_password_authentication", system.AllowSshPasswordAuthentication); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir allow_ssh_password_authentication: %v", err))
	}

	if err := d.Set("allow_multi_factor_authentication", system.AllowMultiFactorAuthentication); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir allow_multi_factor_authentication: %v", err))
	}

	if err := d.Set("description", system.Description); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir description: %v", err))
	}

	// Handle attributes
	if system.Attributes != nil {
		attributes := make(map[string]interface{})
		for k, v := range system.Attributes {
			attributes[k] = fmt.Sprintf("%v", v)
		}
		if err := d.Set("attributes", attributes); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao definir attributes: %v", err))
		}
	}

	return diags
}

func resourceSystemUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	systemID := d.Id()

	// Build system object from resource data
	system := &common.System{
		DisplayName:                    d.Get("display_name").(string),
		AllowSshRootLogin:              d.Get("allow_ssh_root_login").(bool),
		AllowSshPasswordAuthentication: d.Get("allow_ssh_password_authentication").(bool),
		AllowMultiFactorAuthentication: d.Get("allow_multi_factor_authentication").(bool),
		Description:                    d.Get("description").(string),
	}

	// Handle tags if changed
	if tagsRaw, ok := d.GetOk("tags"); ok {
		system.Tags = common.ExpandStringList(tagsRaw.([]interface{}))
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
	c, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	systemID := d.Id()

	// Delete system via API
	_, err = c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/systems/%s", systemID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting system %s: %v", systemID, err))
	}

	// Set ID to empty to signify resource has been removed
	d.SetId("")

	return nil
}
