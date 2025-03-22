package system_groups

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
)

// SystemGroup represents a system group in JumpCloud
type SystemGroup struct {
	ID          string                 `json:"_id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// ResourceGroup returns the resource for managing system groups
func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupCreate,
		ReadContext:   resourceGroupRead,
		UpdateContext: resourceGroupUpdate,
		DeleteContext: resourceGroupDelete,
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

// resourceGroupCreate creates a new system group in JumpCloud
func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Creating system group in JumpCloud")

	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	// Create SystemGroup object from resource data
	group := &SystemGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Type:        "system_group",
	}

	// Process custom attributes if any
	if v, ok := d.GetOk("attributes"); ok {
		group.Attributes = expandAttributes(v.(map[string]interface{}))
	}

	// Convert to JSON
	jsonData, err := json.Marshal(group)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing system group: %v", err))
	}

	// Send request to create the group
	resp, err := client.DoRequest(http.MethodPost, "/api/v2/systemgroups", jsonData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating system group: %v", err))
	}

	// Deserialize the response
	var createdGroup SystemGroup
	if err := json.Unmarshal(resp, &createdGroup); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set resource ID
	d.SetId(createdGroup.ID)

	// Read the resource to update the state
	return resourceGroupRead(ctx, d, meta)
}

// resourceGroupRead reads system group information from JumpCloud
func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Reading system group from JumpCloud")

	var diags diag.Diagnostics

	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	// Get group information by ID
	groupID := d.Id()
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/systemgroups/%s", groupID), nil)
	if err != nil {
		// Check if the group no longer exists
		if isNotFoundError(err) {
			tflog.Warn(ctx, "System group not found, removing from state", map[string]interface{}{
				"id": groupID,
			})
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error fetching system group: %v", err))
	}

	// Deserialize the response
	var group SystemGroup
	if err := json.Unmarshal(resp, &group); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Update resource state
	if err := d.Set("name", group.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description", group.Description); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("attributes", flattenAttributes(group.Attributes)); err != nil {
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
			d.Set("created", metadata.Created.Format(time.RFC3339))
		}
	}

	return diags
}

// resourceGroupUpdate updates an existing system group in JumpCloud
func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Updating system group in JumpCloud")

	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	// Check if there are changes to the fields
	if !d.HasChanges("name", "description", "attributes") {
		return resourceGroupRead(ctx, d, meta)
	}

	// Prepare update object
	group := &SystemGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	// Process custom attributes if any
	if v, ok := d.GetOk("attributes"); ok {
		group.Attributes = expandAttributes(v.(map[string]interface{}))
	}

	// Convert to JSON
	jsonData, err := json.Marshal(group)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing system group: %v", err))
	}

	// Send update request
	groupID := d.Id()
	_, err = client.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/systemgroups/%s", groupID), jsonData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating system group: %v", err))
	}

	// Read the resource to update the state
	return resourceGroupRead(ctx, d, meta)
}

// resourceGroupDelete deletes a system group from JumpCloud
func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Deleting system group from JumpCloud")

	var diags diag.Diagnostics

	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	// Send request to delete the group
	groupID := d.Id()
	_, err := client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/systemgroups/%s", groupID), nil)
	if err != nil {
		// If the resource is already deleted, don't treat as an error
		if isNotFoundError(err) {
			return diags
		}
		return diag.FromErr(fmt.Errorf("error deleting system group: %v", err))
	}

	// Clear the ID to indicate that the resource has been deleted
	d.SetId("")

	return diags
}

// Helper functions

// isNotFoundError checks if the error is a 404 Not Found
func isNotFoundError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found"))
}

// expandAttributes converts map[string]interface{} to map[string]interface{} for the API
func expandAttributes(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range input {
		result[k] = v
	}
	return result
}

// flattenAttributes converts the API response to the Terraform state format
func flattenAttributes(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range input {
		if str, ok := v.(string); ok {
			result[k] = str
		}
	}
	return result
}
