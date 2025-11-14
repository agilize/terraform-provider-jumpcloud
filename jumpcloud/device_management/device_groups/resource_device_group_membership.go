package device_groups

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// ResourceDeviceGroupMembership returns the resource for managing device group memberships
func ResourceDeviceGroupMembership() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDeviceGroupMembershipCreate,
		ReadContext:   resourceDeviceGroupMembershipRead,
		DeleteContext: resourceDeviceGroupMembershipDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"device_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the device group",
			},
			"device_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the device to be associated with the group",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manages the association of devices to device groups in JumpCloud. This resource allows including a device in a specific device group.",
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

// resourceDeviceGroupMembershipCreate creates a new association between a device and a device group
func resourceDeviceGroupMembershipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Creating device to device group association in JumpCloud")

	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	// Map device_group_id and device_id from schema to systemGroupID and systemID for API
	systemGroupID := d.Get("device_group_id").(string)
	systemID := d.Get("device_id").(string)

	// Structure for the request body (API uses "system" terminology)
	requestBody := map[string]interface{}{
		"op":   "add",
		"type": "system",
		"id":   systemID,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing request body: %v", err))
	}

	// Send request to associate the system to the group (API uses "systemgroups")
	_, err = client.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/systemgroups/%s/members", systemGroupID), jsonData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error associating device to group: %v", err))
	}

	// Set the resource ID as a combination of the group and system IDs
	d.SetId(fmt.Sprintf("%s:%s", systemGroupID, systemID))

	return resourceDeviceGroupMembershipRead(ctx, d, meta)
}

// resourceDeviceGroupMembershipRead reads information about an association between a device and a device group
func resourceDeviceGroupMembershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Reading device to device group association from JumpCloud")

	var diags diag.Diagnostics

	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	// Extract IDs from the composite resource ID
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 {
		return diag.FromErr(fmt.Errorf("invalid ID format, expected 'device_group_id:device_id', got: %s", d.Id()))
	}

	systemGroupID := idParts[0]
	systemID := idParts[1]

	// Set attributes in state (map to device_group_id and device_id in schema)
	if err := d.Set("device_group_id", systemGroupID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("device_id", systemID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	// Check if the association still exists (API uses "systemgroups")
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/systemgroups/%s/members", systemGroupID), nil)
	if err != nil {
		// If the group no longer exists, remove from state
		if common.IsNotFoundError(err) {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error fetching group members: %v", err))
	}

	// Decode the response
	var members struct {
		Results []struct {
			To struct {
				ID string `json:"id"`
			} `json:"to"`
		} `json:"results"`
	}
	if err := json.Unmarshal(resp, &members); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Check if the device is still associated with the group
	found := false
	for _, member := range members.Results {
		if member.To.ID == systemID {
			found = true
			break
		}
	}

	// If the device is no longer associated, clear the ID
	if !found {
		d.SetId("")
	}

	return diags
}

// resourceDeviceGroupMembershipDelete removes an association between a device and a device group
func resourceDeviceGroupMembershipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Removing device from device group in JumpCloud")

	var diags diag.Diagnostics

	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	// Extract IDs from the composite resource ID
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 {
		return diag.FromErr(fmt.Errorf("invalid ID format, expected 'device_group_id:device_id', got: %s", d.Id()))
	}

	systemGroupID := idParts[0]
	systemID := idParts[1]

	// Structure for the request body (API uses "system" terminology)
	requestBody := map[string]interface{}{
		"op":   "remove",
		"type": "system",
		"id":   systemID,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing request body: %v", err))
	}

	// Send request to remove the association (API uses "systemgroups")
	_, err = client.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/systemgroups/%s/members", systemGroupID), jsonData)
	if err != nil {
		// Ignore error if the resource has already been removed
		if common.IsNotFoundError(err) {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error removing association: %v", err))
	}

	// Clear the ID to indicate that the resource has been deleted
	d.SetId("")

	return diags
}
