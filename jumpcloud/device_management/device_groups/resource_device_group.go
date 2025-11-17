package device_groups

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

// SystemGroup represents a system group in JumpCloud (API uses "system" terminology)
type SystemGroup struct {
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// ResourceDeviceGroup returns the resource for managing device groups
func ResourceDeviceGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDeviceGroupCreate,
		ReadContext:   resourceDeviceGroupRead,
		UpdateContext: resourceDeviceGroupUpdate,
		DeleteContext: resourceDeviceGroupDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the system group",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the system group",
			},
			"attributes": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Custom attributes of the system group",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the system group",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manages system groups in JumpCloud. This resource allows creating, updating and deleting system groups, facilitating organization and management of systems.",
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Update: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

// resourceDeviceGroupCreate creates a new device group in JumpCloud
func resourceDeviceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Creating device group in JumpCloud")

	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	// Create SystemGroup object from resource data (API uses "system" terminology)
	group := &SystemGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Type:        "system_group",
	}

	// Handle attributes
	if v, ok := d.GetOk("attributes"); ok {
		attributesMap := v.(map[string]interface{})
		group.Attributes = common.ExpandAttributes(attributesMap)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(group)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing system group: %v", err))
	}

	// Send request to create the group (API uses "systemgroups")
	resp, err := client.DoRequest(http.MethodPost, "/api/v2/systemgroups", jsonData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating system group: %v", err))
	}

	// Log the raw response for debugging
	tflog.Debug(ctx, fmt.Sprintf("Create device group API response: %s", string(resp)))

	// Try to decode response as a single object first
	var createdGroup SystemGroup
	if err := json.Unmarshal(resp, &createdGroup); err != nil {
		tflog.Debug(ctx, fmt.Sprintf("Failed to unmarshal as single object: %v", err))

		// If that fails, try to decode as an array
		var groups []SystemGroup
		if err2 := json.Unmarshal(resp, &groups); err2 != nil {
			tflog.Error(ctx, fmt.Sprintf("Failed to unmarshal as array: %v", err2))
			tflog.Error(ctx, fmt.Sprintf("Raw response was: %s", string(resp)))
			return diag.FromErr(fmt.Errorf("error deserializing response: %v, %v", err, err2))
		}

		// Use the first group in the array
		if len(groups) == 0 {
			return diag.FromErr(fmt.Errorf("no groups returned from API"))
		}
		tflog.Debug(ctx, fmt.Sprintf("Successfully unmarshaled as array with %d groups", len(groups)))
		createdGroup = groups[0]
	} else {
		tflog.Debug(ctx, "Successfully unmarshaled as single object")
	}

	// Set resource ID
	tflog.Debug(ctx, fmt.Sprintf("Setting device group ID to: '%s'", createdGroup.ID))
	d.SetId(createdGroup.ID)
	tflog.Debug(ctx, fmt.Sprintf("Device group ID after SetId: '%s'", d.Id()))

	// Read the resource to update the state
	return resourceDeviceGroupRead(ctx, d, meta)
}

// resourceDeviceGroupRead reads device group information from JumpCloud
func resourceDeviceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Reading device group from JumpCloud")

	var diags diag.Diagnostics

	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	// Get group information by ID (API uses "systemgroups")
	groupID := d.Id()
	tflog.Debug(ctx, fmt.Sprintf("Reading device group with ID: '%s'", groupID))

	if groupID == "" {
		return diag.FromErr(fmt.Errorf("device group ID is empty"))
	}

	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/systemgroups/%s", groupID), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Device group %s not found, removing from state", groupID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading device group: %v", err))
	}

	// Log the raw response for debugging
	tflog.Debug(ctx, fmt.Sprintf("Read device group API response: %s", string(resp)))

	// Deserialize the response - try single object first, then array
	var group SystemGroup
	if err := json.Unmarshal(resp, &group); err != nil {
		tflog.Debug(ctx, fmt.Sprintf("Failed to unmarshal as single object: %v", err))

		// If that fails, try to decode as an array
		var groups []SystemGroup
		if err2 := json.Unmarshal(resp, &groups); err2 != nil {
			tflog.Error(ctx, fmt.Sprintf("Failed to unmarshal as array: %v", err2))
			tflog.Error(ctx, fmt.Sprintf("Raw response was: %s", string(resp)))
			return diag.FromErr(fmt.Errorf("error deserializing response: %v, %v", err, err2))
		}

		// Find the group with the matching ID in the array
		if len(groups) == 0 {
			return diag.FromErr(fmt.Errorf("no groups returned from API"))
		}
		tflog.Debug(ctx, fmt.Sprintf("Successfully unmarshaled as array with %d groups, looking for ID: %s", len(groups), groupID))

		found := false
		for _, g := range groups {
			if g.ID == groupID {
				group = g
				found = true
				tflog.Debug(ctx, fmt.Sprintf("Found matching group in array: %s", g.Name))
				break
			}
		}

		if !found {
			return diag.FromErr(fmt.Errorf("group with ID %s not found in API response", groupID))
		}
	} else {
		tflog.Debug(ctx, "Successfully unmarshaled as single object")
	}

	// Update resource state
	if err := d.Set("name", group.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description", group.Description); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("attributes", common.FlattenAttributes(group.Attributes)); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	// Get additional group metadata
	metaResp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/systemgroups/%s/members", groupID), nil)
	if err == nil {
		var metadata struct {
			TotalCount int       `json:"totalCount"`
			Created    time.Time `json:"created"`
		}
		if err := json.Unmarshal(metaResp, &metadata); err == nil {
			if err := d.Set("created", metadata.Created.Format(time.RFC3339)); err != nil {
				return diag.FromErr(fmt.Errorf("error setting created: %v", err))
			}
		}
	}

	return diags
}

// resourceDeviceGroupUpdate updates an existing device group in JumpCloud
func resourceDeviceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Updating device group in JumpCloud")

	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	// Check if there are changes to the fields
	if !d.HasChanges("name", "description", "attributes") {
		return resourceDeviceGroupRead(ctx, d, meta)
	}

	// Prepare update object
	group := &SystemGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	// Handle attributes
	if v, ok := d.GetOk("attributes"); ok {
		attributesMap := v.(map[string]interface{})
		group.Attributes = common.ExpandAttributes(attributesMap)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(group)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing system group: %v", err))
	}

	// Send update request (API uses "systemgroups")
	groupID := d.Id()
	_, err = client.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/systemgroups/%s", groupID), jsonData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating device group: %v", err))
	}

	// Read the resource to update the state
	return resourceDeviceGroupRead(ctx, d, meta)
}

// resourceDeviceGroupDelete deletes a device group from JumpCloud
func resourceDeviceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Deleting device group from JumpCloud")

	var diags diag.Diagnostics

	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	// Send request to delete the group (API uses "systemgroups")
	groupID := d.Id()
	_, err := client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/systemgroups/%s", groupID), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Device group %s not found, assuming already deleted", groupID))
			return diags
		}
		return diag.FromErr(fmt.Errorf("error deleting device group: %v", err))
	}

	// Clear the ID to indicate that the resource has been deleted
	d.SetId("")

	return diags
}
