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

func ResourceDevice() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDeviceCreate,
		ReadContext:   resourceDeviceRead,
		UpdateContext: resourceDeviceUpdate,
		DeleteContext: resourceDeviceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manages an existing JumpCloud device. Note: Devices cannot be created via the API - they must be registered by installing the JumpCloud agent. Use 'terraform import' to manage existing devices.",
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique identifier of the device.",
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The display name of the device. If not specified, the current name from JumpCloud will be used.",
			},
			"device_type": {
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

func resourceDeviceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Devices cannot be created via the API - they must be registered by installing the JumpCloud agent
	return diag.Errorf(`Devices cannot be created via the API.

To manage an existing device with Terraform:

1. Install the JumpCloud agent on your device
2. Find the device in JumpCloud console or use the data source:

   data "jumpcloud_devices" "my_device" {
     display_name = "DEVICE-NAME"
   }

3. Import the device into Terraform:

   import {
     to = jumpcloud_devices.my_device
     id = data.jumpcloud_devices.my_device.id
   }

4. Define the resource configuration:

   resource "jumpcloud_devices" "my_device" {
     tags = ["managed", "production"]
     allow_ssh_password_authentication = false
     # ... other configuration
   }

For more information, see: https://docs.jumpcloud.com/`)
}

func resourceDeviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("device ID not provided"))
	}

	// Fetch device via API (uses API v1 and "system" terminology internally)
	tflog.Debug(ctx, fmt.Sprintf("Reading device with ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/systems/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Device %s not found, removing from state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading device: %v", err))
	}

	// Deserialize response
	var system common.System
	if err := json.Unmarshal(resp, &system); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set tags if they exist
	if len(system.Tags) > 0 {
		if err := d.Set("tags", common.FlattenStringList(system.Tags)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting tags: %v", err))
		}
	}

	// Set fields in resource data
	if err := d.Set("display_name", system.DisplayName); err != nil {
		return diag.FromErr(fmt.Errorf("error setting display_name: %v", err))
	}

	// Map system_type from API to device_type in schema
	if err := d.Set("device_type", system.SystemType); err != nil {
		return diag.FromErr(fmt.Errorf("error setting device_type: %v", err))
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

	// Handle attributes
	if len(system.Attributes) > 0 {
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

func resourceDeviceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	deviceID := d.Id()

	tflog.Debug(ctx, fmt.Sprintf("Updating device with ID: %s", deviceID))

	// Build system object from resource data (API uses "system" terminology internally)
	system := &common.System{
		AllowSshRootLogin:              d.Get("allow_ssh_root_login").(bool),
		AllowSshPasswordAuthentication: d.Get("allow_ssh_password_authentication").(bool),
		AllowMultiFactorAuthentication: d.Get("allow_multi_factor_authentication").(bool),
		Description:                    d.Get("description").(string),
	}

	// Only set display_name if it was explicitly provided
	if displayName, ok := d.GetOk("display_name"); ok {
		system.DisplayName = displayName.(string)
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
		return diag.FromErr(fmt.Errorf("error serializing device: %v", err))
	}

	tflog.Debug(ctx, fmt.Sprintf("Updating device %s with data: %s", deviceID, string(systemJSON)))

	// Update device via API (uses API v1)
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/systems/%s", deviceID), systemJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating device %s: %v", deviceID, err))
	}

	return resourceDeviceRead(ctx, d, meta)
}

func resourceDeviceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	deviceID := d.Id()

	tflog.Debug(ctx, fmt.Sprintf("Deleting device with ID: %s", deviceID))

	// Delete device via API (uses API v1)
	_, err = c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/systems/%s", deviceID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting device %s: %v", deviceID, err))
	}

	tflog.Info(ctx, fmt.Sprintf("Device %s deleted successfully", deviceID))

	// Set ID to empty to signify resource has been removed
	d.SetId("")

	return nil
}
