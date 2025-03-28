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
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// ResourceUserMapping defines the resource for managing user mappings in applications
func ResourceUserMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserMappingCreate,
		ReadContext:   resourceUserMappingRead,
		UpdateContext: resourceUserMappingUpdate,
		DeleteContext: resourceUserMappingDelete,

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "JumpCloud application ID",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "JumpCloud user ID",
			},
			"attributes": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Additional attributes for the user mapping (application-specific)",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceUserMappingImport,
		},
	}
}

// resourceUserMappingCreate creates a new user mapping for an application
func resourceUserMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get IDs
	applicationID := d.Get("application_id").(string)
	userID := d.Get("user_id").(string)

	// Parameter validation
	if applicationID == "" {
		return diag.FromErr(fmt.Errorf("application_id cannot be empty"))
	}

	if userID == "" {
		return diag.FromErr(fmt.Errorf("user_id cannot be empty"))
	}

	// Create mapping structure
	mapping := &UserMapping{
		ApplicationID: applicationID,
		UserID:        userID,
	}

	// Process attributes
	if attrs, ok := d.GetOk("attributes"); ok {
		attrMap := attrs.(map[string]interface{})
		mapping.Attributes = common.ExpandAttributes(attrMap)
	}

	// Serialize mapping to JSON
	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing mapping: %v", err))
	}

	// Call API to create mapping
	tflog.Debug(ctx, fmt.Sprintf("Creating mapping between application %s and user %s", applicationID, userID))
	resp, err := client.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/applications/%s/users", applicationID), mappingJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating user mapping: %v", err))
	}

	// Deserialize response
	var createdMapping UserMapping
	if err := json.Unmarshal(resp, &createdMapping); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// If API doesn't return a specific ID for the mapping, use the combination of IDs
	if createdMapping.ID == "" {
		d.SetId(fmt.Sprintf("%s:%s", applicationID, userID))
	} else {
		d.SetId(createdMapping.ID)
	}

	return resourceUserMappingRead(ctx, d, meta)
}

// resourceUserMappingRead reads user mapping data
func resourceUserMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Extract IDs from resource ID if it's a composite ID
	var applicationID, userID string

	if strings.Contains(d.Id(), ":") {
		parts := strings.Split(d.Id(), ":")
		if len(parts) != 2 {
			return diag.FromErr(fmt.Errorf("invalid ID: %s. Expected format: {application_id}:{user_id}", d.Id()))
		}
		applicationID = parts[0]
		userID = parts[1]
	} else {
		// Using values from state if available
		applicationID = d.Get("application_id").(string)
		userID = d.Get("user_id").(string)
	}

	// Call API to fetch all user mappings for the application
	tflog.Debug(ctx, fmt.Sprintf("Fetching user mappings for application %s", applicationID))
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/applications/%s/users", applicationID), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Application %s not found, removing from state", applicationID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error fetching user mappings: %v", err))
	}

	// Deserialize response
	var mappings []UserMapping
	if err := json.Unmarshal(resp, &mappings); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Look for the specific mapping
	found := false
	var mapping UserMapping

	for _, m := range mappings {
		if m.UserID == userID {
			mapping = m
			found = true
			break
		}
	}

	if !found {
		tflog.Warn(ctx, fmt.Sprintf("Mapping between application %s and user %s not found, removing from state", applicationID, userID))
		d.SetId("")
		return diags
	}

	// Update state
	d.Set("application_id", applicationID)
	d.Set("user_id", userID)

	// Process attributes if present
	if mapping.Attributes != nil {
		// Convert attributes to a map that can be stored in the Terraform state
		attributeMap := make(map[string]interface{})

		for k, v := range mapping.Attributes {
			if str, ok := v.(string); ok {
				attributeMap[k] = str
			} else {
				// For non-string values, convert to JSON
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					return diag.FromErr(fmt.Errorf("erro ao converter atributo para JSON: %v", err))
				}
				attributeMap[k] = string(jsonBytes)
			}
		}

		d.Set("attributes", attributeMap)
	}

	return diags
}

// resourceUserMappingUpdate updates a user mapping
func resourceUserMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get IDs
	applicationID := d.Get("application_id").(string)
	userID := d.Get("user_id").(string)

	// If attributes haven't changed, no need to update
	if !d.HasChange("attributes") {
		return resourceUserMappingRead(ctx, d, meta)
	}

	// Create updated mapping structure
	mapping := &UserMapping{
		ApplicationID: applicationID,
		UserID:        userID,
	}

	// Process attributes
	if attrs, ok := d.GetOk("attributes"); ok {
		attrMap := attrs.(map[string]interface{})
		mapping.Attributes = common.ExpandAttributes(attrMap)
	}

	// Serialize mapping to JSON
	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing mapping: %v", err))
	}

	// Call API to update mapping
	tflog.Debug(ctx, fmt.Sprintf("Updating mapping between application %s and user %s", applicationID, userID))
	_, err = client.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/applications/%s/users/%s", applicationID, userID), mappingJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating user mapping: %v", err))
	}

	return resourceUserMappingRead(ctx, d, meta)
}

// resourceUserMappingDelete removes a user mapping
func resourceUserMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get IDs
	applicationID := d.Get("application_id").(string)
	userID := d.Get("user_id").(string)

	// Call API to delete mapping
	tflog.Debug(ctx, fmt.Sprintf("Removing mapping between application %s and user %s", applicationID, userID))
	_, err = client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/applications/%s/users/%s", applicationID, userID), nil)
	if err != nil {
		// If the resource is already gone, just log a warning
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Mapping between application %s and user %s not found or already deleted", applicationID, userID))
		} else {
			return diag.FromErr(fmt.Errorf("error deleting user mapping: %v", err))
		}
	}

	// Remove ID from state
	d.SetId("")

	return diags
}

// resourceUserMappingImport imports an existing mapping
func resourceUserMappingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Expected format: {application_id}:{user_id}
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ID format, use: {application_id}:{user_id}")
	}

	applicationID := parts[0]
	userID := parts[1]

	d.SetId(fmt.Sprintf("%s:%s", applicationID, userID))
	d.Set("application_id", applicationID)
	d.Set("user_id", userID)

	return []*schema.ResourceData{d}, nil
}
