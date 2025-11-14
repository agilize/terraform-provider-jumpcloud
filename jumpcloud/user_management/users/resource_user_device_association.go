package users

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

// ResourceUserDeviceAssociation returns the resource for managing associations between users and devices
func ResourceUserDeviceAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserDeviceAssociationCreate,
		ReadContext:   resourceUserDeviceAssociationRead,
		DeleteContext: resourceUserDeviceAssociationDelete,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the JumpCloud user",
			},
			"device_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the JumpCloud device",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceUserDeviceAssociationImport,
		},
		Description: "Manages the association between a user and a device in JumpCloud.",
	}
}

// resourceUserDeviceAssociationCreate creates an association between a user and a system
func resourceUserDeviceAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Creating user-system association in JumpCloud")

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	userID := d.Get("user_id").(string)
	systemID := d.Get("device_id").(string) // Map device_id from schema to systemID for API

	// Parameter validation
	if userID == "" {
		return diag.FromErr(fmt.Errorf("user_id cannot be empty"))
	}

	if systemID == "" {
		return diag.FromErr(fmt.Errorf("device_id cannot be empty"))
	}

	// In JumpCloud, the API to associate a user with a system is:
	// POST /api/v2/users/{user_id}/systems/{system_id}
	tflog.Debug(ctx, fmt.Sprintf("Creating association between user %s and system %s", userID, systemID))
	_, err = client.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/users/%s/systems/%s", userID, systemID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating user-system association: %v", err))
	}

	// The association ID is a combination of the user and system IDs
	d.SetId(fmt.Sprintf("%s:%s", userID, systemID))

	return resourceUserDeviceAssociationRead(ctx, d, meta)
}

// resourceUserDeviceAssociationRead reads information about a user-system association
func resourceUserDeviceAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.Debug(ctx, fmt.Sprintf("Reading user-system association in JumpCloud: %s", d.Id()))

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Extract user_id and system_id from the association ID
	userID, systemID, err := parseSystemAssociationID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Verify if the association exists
	// GET /api/v2/users/{user_id}/systems
	tflog.Debug(ctx, fmt.Sprintf("Fetching systems for user %s", userID))
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/users/%s/systems", userID), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("User %s not found, removing association from state", userID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error checking user-system association: %v", err))
	}

	// Check if the systemID is in the response
	var systems []struct {
		ID string `json:"_id"`
	}
	if err := json.Unmarshal(resp, &systems); err != nil {
		return diag.FromErr(fmt.Errorf("error decoding response: %v", err))
	}

	found := false
	for _, system := range systems {
		if system.ID == systemID {
			found = true
			break
		}
	}

	if !found {
		// If the association doesn't exist, clear the state
		tflog.Warn(ctx, fmt.Sprintf("Association between user %s and system %s not found, removing from state", userID, systemID))
		d.SetId("")
		return diags
	}

	if err := d.Set("user_id", userID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_id: %v", err))
	}
	// Map systemID from API to device_id in schema
	if err := d.Set("device_id", systemID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting device_id: %v", err))
	}

	return diags
}

// resourceUserDeviceAssociationDelete removes an association between a user and a system
func resourceUserDeviceAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.Debug(ctx, fmt.Sprintf("Removing user-system association in JumpCloud: %s", d.Id()))

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Extract user_id and system_id from the association ID
	userID, systemID, err := parseSystemAssociationID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// In JumpCloud, the API to remove an association is:
	// DELETE /api/v2/users/{user_id}/systems/{system_id}
	tflog.Debug(ctx, fmt.Sprintf("Removing association between user %s and system %s", userID, systemID))
	_, err = client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/users/%s/systems/%s", userID, systemID), nil)
	if err != nil {
		// If the resource is already gone, just log a warning
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Association between user %s and system %s not found or already deleted", userID, systemID))
		} else {
			return diag.FromErr(fmt.Errorf("error removing user-system association: %v", err))
		}
	}

	// Clear the resource ID
	d.SetId("")

	return diags
}

// parseSystemAssociationID extracts user_id and system_id from the association ID
func parseSystemAssociationID(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid format for user-system association ID: %s, expected format 'user_id:system_id'", id)
	}

	userID := parts[0]
	systemID := parts[1]

	if userID == "" || systemID == "" {
		return "", "", fmt.Errorf("user_id and system_id cannot be empty in the association ID")
	}

	return userID, systemID, nil
}

// resourceUserDeviceAssociationImport imports an existing user-system association
func resourceUserDeviceAssociationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Expected format: {user_id}:{system_id}
	userID, systemID, err := parseSystemAssociationID(d.Id())
	if err != nil {
		return nil, err
	}

	d.SetId(fmt.Sprintf("%s:%s", userID, systemID))

	if err := d.Set("user_id", userID); err != nil {
		return nil, fmt.Errorf("error setting user_id: %v", err)
	}

	if err := d.Set("system_id", systemID); err != nil {
		return nil, fmt.Errorf("error setting system_id: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}
