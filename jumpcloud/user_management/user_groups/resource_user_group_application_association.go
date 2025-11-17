package usergroups

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

// GraphAssociationRequest represents the request body for Graph API associations
type GraphAssociationRequest struct {
	Op   string `json:"op"`   // "add" or "remove"
	Type string `json:"type"` // "application"
	ID   string `json:"id"`   // application ID
}

// GraphAssociationResponse represents an association in the Graph API response
type GraphAssociationResponse struct {
	ID                 string                 `json:"id"`
	Type               string                 `json:"type"`
	CompiledAttributes map[string]interface{} `json:"compiledAttributes,omitempty"`
	Paths              []interface{}          `json:"paths,omitempty"`
}

// ResourceUserGroupApplicationAssociation defines the resource for managing user group to application associations
func ResourceUserGroupApplicationAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserGroupApplicationAssociationCreate,
		ReadContext:   resourceUserGroupApplicationAssociationRead,
		DeleteContext: resourceUserGroupApplicationAssociationDelete,

		Schema: map[string]*schema.Schema{
			"user_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "JumpCloud user group ID",
			},
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "JumpCloud application ID",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceUserGroupApplicationAssociationImport,
		},
	}
}

// resourceUserGroupApplicationAssociationCreate creates a new user group to application association
func resourceUserGroupApplicationAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	userGroupID := d.Get("user_group_id").(string)
	applicationID := d.Get("application_id").(string)

	// Parameter validation
	if userGroupID == "" {
		return diag.FromErr(fmt.Errorf("user_group_id cannot be empty"))
	}

	if applicationID == "" {
		return diag.FromErr(fmt.Errorf("application_id cannot be empty"))
	}

	// Create association request using Graph API format
	associationReq := &GraphAssociationRequest{
		Op:   "add",
		Type: "application",
		ID:   applicationID,
	}

	// Serialize to JSON
	reqJSON, err := json.Marshal(associationReq)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing association request: %v", err))
	}

	// Use Graph API endpoint
	endpoint := fmt.Sprintf("/api/v2/usergroups/%s/associations", userGroupID)

	// Call API to create association
	tflog.Debug(ctx, fmt.Sprintf("Creating association between user group %s and application %s", userGroupID, applicationID))
	_, err = client.DoRequest(http.MethodPost, endpoint, reqJSON)
	if err != nil {
		// Check if it's a 409 Conflict (already exists)
		if strings.Contains(err.Error(), "409") || strings.Contains(err.Error(), "Conflict") {
			tflog.Warn(ctx, fmt.Sprintf("Association already exists between user group %s and application %s", userGroupID, applicationID))
			// Set ID and continue - the association already exists
			d.SetId(fmt.Sprintf("%s:%s", userGroupID, applicationID))
			return resourceUserGroupApplicationAssociationRead(ctx, d, meta)
		}
		return diag.FromErr(fmt.Errorf("error creating user group application association: %v", err))
	}

	// Set resource ID as composite of user_group_id:application_id
	d.SetId(fmt.Sprintf("%s:%s", userGroupID, applicationID))

	return resourceUserGroupApplicationAssociationRead(ctx, d, meta)
}

// resourceUserGroupApplicationAssociationRead reads user group application association data
func resourceUserGroupApplicationAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Extract IDs from resource ID if it's a composite ID
	var userGroupID, applicationID string

	if strings.Contains(d.Id(), ":") {
		parts := strings.Split(d.Id(), ":")
		if len(parts) != 2 {
			return diag.FromErr(fmt.Errorf("invalid ID: %s. Expected format: {user_group_id}:{application_id}", d.Id()))
		}
		userGroupID = parts[0]
		applicationID = parts[1]
	} else {
		// Using values from state if available
		userGroupID = d.Get("user_group_id").(string)
		applicationID = d.Get("application_id").(string)
	}

	// Use Graph API endpoint to get associations
	endpoint := fmt.Sprintf("/api/v2/usergroups/%s/applications", userGroupID)

	// Call API to fetch all application associations for the user group
	tflog.Debug(ctx, fmt.Sprintf("Fetching application associations for user group %s", userGroupID))
	resp, err := client.DoRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("User group %s not found, removing from state", userGroupID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error fetching application associations: %v", err))
	}

	// Deserialize response
	var associations []GraphAssociationResponse
	if err := json.Unmarshal(resp, &associations); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Look for the specific association
	found := false
	for _, assoc := range associations {
		if assoc.ID == applicationID && assoc.Type == "application" {
			found = true
			break
		}
	}

	if !found {
		tflog.Warn(ctx, fmt.Sprintf("Association between user group %s and application %s not found, removing from state", userGroupID, applicationID))
		d.SetId("")
		return diags
	}

	// Update state
	if err := d.Set("user_group_id", userGroupID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_group_id: %v", err))
	}

	if err := d.Set("application_id", applicationID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting application_id: %v", err))
	}

	return diags
}

// resourceUserGroupApplicationAssociationDelete removes a user group to application association
func resourceUserGroupApplicationAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	userGroupID := d.Get("user_group_id").(string)
	applicationID := d.Get("application_id").(string)

	// Create association request using Graph API format
	associationReq := &GraphAssociationRequest{
		Op:   "remove",
		Type: "application",
		ID:   applicationID,
	}

	// Serialize to JSON
	reqJSON, err := json.Marshal(associationReq)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing association request: %v", err))
	}

	// Use Graph API endpoint
	endpoint := fmt.Sprintf("/api/v2/usergroups/%s/associations", userGroupID)

	// Call API to delete association
	tflog.Debug(ctx, fmt.Sprintf("Removing association between user group %s and application %s", userGroupID, applicationID))
	_, err = client.DoRequest(http.MethodPost, endpoint, reqJSON)
	if err != nil {
		// If the resource is already gone, just log a warning
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Association between user group %s and application %s not found or already deleted", userGroupID, applicationID))
		} else {
			return diag.FromErr(fmt.Errorf("error deleting user group application association: %v", err))
		}
	}

	// Remove ID from state
	d.SetId("")

	return diags
}

// resourceUserGroupApplicationAssociationImport imports an existing association
func resourceUserGroupApplicationAssociationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Expected format: {user_group_id}:{application_id}
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ID format, use: {user_group_id}:{application_id}")
	}

	userGroupID := parts[0]
	applicationID := parts[1]

	d.SetId(fmt.Sprintf("%s:%s", userGroupID, applicationID))

	if err := d.Set("user_group_id", userGroupID); err != nil {
		return nil, fmt.Errorf("error setting user_group_id: %v", err)
	}

	if err := d.Set("application_id", applicationID); err != nil {
		return nil, fmt.Errorf("error setting application_id: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}
