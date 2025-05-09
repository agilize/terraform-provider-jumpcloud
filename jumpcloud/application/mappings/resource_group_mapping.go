package mappings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// expandAttributes converts Terraform map to a format suitable for API
func expandAttributes(attrs map[string]interface{}) map[string]interface{} {
	if attrs == nil {
		return nil
	}

	result := make(map[string]interface{})
	for k, v := range attrs {
		result[k] = v
	}
	return result
}

// flattenAttributes converts API attributes to a format suitable for Terraform state
func flattenAttributes(attrs map[string]interface{}) map[string]interface{} {
	if attrs == nil {
		return nil
	}

	result := make(map[string]interface{})
	for k, v := range attrs {
		// Convert values to strings for Terraform state
		switch val := v.(type) {
		case string:
			result[k] = val
		case bool:
			result[k] = fmt.Sprintf("%t", val)
		case float64:
			result[k] = fmt.Sprintf("%g", val)
		case json.Number:
			result[k] = val.String()
		case map[string]interface{}:
			// If complex nested structure, convert to JSON
			jsonBytes, err := json.Marshal(val)
			if err == nil {
				result[k] = string(jsonBytes)
			} else {
				result[k] = fmt.Sprintf("%v", val)
			}
		default:
			result[k] = fmt.Sprintf("%v", val)
		}
	}
	return result
}

// ResourceGroupMapping defines the resource for managing group mappings in applications
func ResourceGroupMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupMappingCreate,
		ReadContext:   resourceGroupMappingRead,
		UpdateContext: resourceGroupMappingUpdate,
		DeleteContext: resourceGroupMappingDelete,

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "JumpCloud application ID",
			},
			"group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "JumpCloud user group ID",
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "user_group",
				ValidateFunc: validation.StringInSlice([]string{"user_group", "system_group"}, false),
				Description:  "Group type: 'user_group' (default) or 'system_group'",
				ForceNew:     true,
			},
			"attributes": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Additional attributes for the group mapping (application-specific)",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceGroupMappingImport,
		},
	}
}

// resourceGroupMappingCreate creates a new group mapping for an application
func resourceGroupMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get IDs and type
	applicationID := d.Get("application_id").(string)
	groupID := d.Get("group_id").(string)
	groupType := d.Get("type").(string)

	// Parameter validation
	if applicationID == "" {
		return diag.FromErr(fmt.Errorf("application_id cannot be empty"))
	}

	if groupID == "" {
		return diag.FromErr(fmt.Errorf("group_id cannot be empty"))
	}

	// Create mapping structure
	mapping := &GroupMapping{
		ApplicationID: applicationID,
		GroupID:       groupID,
		Type:          groupType,
	}

	// Include additional attributes if present
	if v, ok := d.GetOk("attributes"); ok {
		mapping.Attributes = expandAttributes(v.(map[string]interface{}))
	}

	// Serialize mapping to JSON
	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing mapping: %v", err))
	}

	// Determine the correct endpoint based on group type
	var endpoint string
	if groupType == "system_group" {
		endpoint = fmt.Sprintf("/api/v2/applications/%s/systemgroups", applicationID)
	} else {
		endpoint = fmt.Sprintf("/api/v2/applications/%s/usergroups", applicationID)
	}

	// Call API to create mapping
	tflog.Debug(ctx, fmt.Sprintf("Creating mapping between application %s and group %s of type %s", applicationID, groupID, groupType))
	resp, err := client.DoRequest(http.MethodPost, endpoint, mappingJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating group mapping: %v", err))
	}

	// Deserialize response
	var createdMapping GroupMapping
	if err := json.Unmarshal(resp, &createdMapping); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// If API doesn't return a specific ID for the mapping, use the combination of IDs
	if createdMapping.ID == "" {
		d.SetId(fmt.Sprintf("%s:%s:%s", applicationID, groupType, groupID))
	} else {
		d.SetId(createdMapping.ID)
	}

	return resourceGroupMappingRead(ctx, d, meta)
}

// resourceGroupMappingRead reads group mapping data
func resourceGroupMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Extract IDs from resource ID if it's a composite ID
	var applicationID, groupID, groupType string

	if strings.Contains(d.Id(), ":") {
		parts := strings.Split(d.Id(), ":")
		if len(parts) != 3 {
			return diag.FromErr(fmt.Errorf("invalid ID: %s. Expected format: {application_id}:{group_type}:{group_id}", d.Id()))
		}
		applicationID = parts[0]
		groupType = parts[1]
		groupID = parts[2]
	} else {
		// Using values from state if available
		applicationID = d.Get("application_id").(string)
		groupType = d.Get("type").(string)
		groupID = d.Get("group_id").(string)
	}

	// Determine the correct endpoint based on group type
	var endpoint string
	if groupType == "system_group" {
		endpoint = fmt.Sprintf("/api/v2/applications/%s/systemgroups", applicationID)
	} else {
		endpoint = fmt.Sprintf("/api/v2/applications/%s/usergroups", applicationID)
	}

	// Call API to fetch all group mappings for the application
	tflog.Debug(ctx, fmt.Sprintf("Fetching group mappings for application %s", applicationID))
	resp, err := client.DoRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Application %s not found, removing from state", applicationID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error fetching group mappings: %v", err))
	}

	// Deserialize response
	var mappings []GroupMapping
	if err := json.Unmarshal(resp, &mappings); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Look for the specific mapping
	found := false
	var mapping GroupMapping

	for _, m := range mappings {
		if m.GroupID == groupID {
			mapping = m
			found = true
			break
		}
	}

	if !found {
		tflog.Warn(ctx, fmt.Sprintf("Mapping between application %s and group %s of type %s not found, removing from state", applicationID, groupID, groupType))
		d.SetId("")
		return diags
	}

	// Update state
	if err := d.Set("application_id", applicationID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting application_id: %v", err))
	}

	if err := d.Set("group_id", groupID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting group_id: %v", err))
	}

	if err := d.Set("type", groupType); err != nil {
		return diag.FromErr(fmt.Errorf("error setting type: %v", err))
	}

	// Update attributes
	if mapping.Attributes != nil {
		if err := d.Set("attributes", flattenAttributes(mapping.Attributes)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting attributes: %v", err))
		}
	}

	return diags
}

// resourceGroupMappingUpdate updates a group mapping
func resourceGroupMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get IDs and type
	applicationID := d.Get("application_id").(string)
	groupID := d.Get("group_id").(string)
	groupType := d.Get("type").(string)

	// If attributes haven't changed, no need to update
	if !d.HasChange("attributes") {
		return resourceGroupMappingRead(ctx, d, meta)
	}

	// Create updated mapping structure
	mapping := &GroupMapping{
		ApplicationID: applicationID,
		GroupID:       groupID,
		Type:          groupType,
	}

	// Include updated attributes
	if v, ok := d.GetOk("attributes"); ok {
		mapping.Attributes = expandAttributes(v.(map[string]interface{}))
	}

	// Serialize mapping to JSON
	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing mapping: %v", err))
	}

	// Determine the correct endpoint based on group type
	var endpoint string
	if groupType == "system_group" {
		endpoint = fmt.Sprintf("/api/v2/applications/%s/systemgroups/%s", applicationID, groupID)
	} else {
		endpoint = fmt.Sprintf("/api/v2/applications/%s/usergroups/%s", applicationID, groupID)
	}

	// Call API to update mapping
	tflog.Debug(ctx, fmt.Sprintf("Updating mapping between application %s and group %s of type %s", applicationID, groupID, groupType))
	_, err = client.DoRequest(http.MethodPut, endpoint, mappingJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating group mapping: %v", err))
	}

	return resourceGroupMappingRead(ctx, d, meta)
}

// resourceGroupMappingDelete removes a group mapping
func resourceGroupMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get IDs and type
	applicationID := d.Get("application_id").(string)
	groupID := d.Get("group_id").(string)
	groupType := d.Get("type").(string)

	// Determine the correct endpoint based on group type
	var endpoint string
	if groupType == "system_group" {
		endpoint = fmt.Sprintf("/api/v2/applications/%s/systemgroups/%s", applicationID, groupID)
	} else {
		endpoint = fmt.Sprintf("/api/v2/applications/%s/usergroups/%s", applicationID, groupID)
	}

	// Call API to delete mapping
	tflog.Debug(ctx, fmt.Sprintf("Removing mapping between application %s and group %s of type %s", applicationID, groupID, groupType))
	_, err = client.DoRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		// If the resource is already gone, just log a warning
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Mapping between application %s and group %s of type %s not found or already deleted", applicationID, groupID, groupType))
		} else {
			return diag.FromErr(fmt.Errorf("error deleting group mapping: %v", err))
		}
	}

	// Remove ID from state
	d.SetId("")

	return diags
}

// resourceGroupMappingImport imports an existing mapping
func resourceGroupMappingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Expected format: {application_id}:{group_type}:{group_id}
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid ID format, use: {application_id}:{group_type}:{group_id}")
	}

	applicationID := parts[0]
	groupType := parts[1]
	groupID := parts[2]

	// Validate group type
	if groupType != "user_group" && groupType != "system_group" {
		return nil, fmt.Errorf("invalid group type: %s, must be either 'user_group' or 'system_group'", groupType)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", applicationID, groupType, groupID))

	if err := d.Set("application_id", applicationID); err != nil {
		return nil, fmt.Errorf("error setting application_id: %v", err)
	}

	if err := d.Set("type", groupType); err != nil {
		return nil, fmt.Errorf("error setting type: %v", err)
	}

	if err := d.Set("group_id", groupID); err != nil {
		return nil, fmt.Errorf("error setting group_id: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}
