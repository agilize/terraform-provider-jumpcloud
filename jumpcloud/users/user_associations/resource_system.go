package user_associations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ResourceSystem returns the resource for managing associations between users and systems
func ResourceSystem() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSystemCreate,
		ReadContext:   resourceSystemRead,
		DeleteContext: resourceSystemDelete,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the JumpCloud user",
			},
			"system_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the JumpCloud system",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manages the association between a user and a system in JumpCloud.",
	}
}

// resourceSystemCreate creates an association between a user and a system
func resourceSystemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Creating user-system association in JumpCloud")

	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	userID := d.Get("user_id").(string)
	systemID := d.Get("system_id").(string)

	// In JumpCloud, the API to associate a user with a system is:
	// POST /api/v2/users/{user_id}/systems/{system_id}
	_, err := client.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/users/%s/systems/%s", userID, systemID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating user-system association: %v", err))
	}

	// The association ID is a combination of the user and system IDs
	d.SetId(fmt.Sprintf("%s:%s", userID, systemID))

	return resourceSystemRead(ctx, d, meta)
}

// resourceSystemRead reads information about a user-system association
func resourceSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Reading user-system association in JumpCloud: %s", d.Id()))

	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	// Extract user_id and system_id from the association ID
	userID, systemID, err := parseSystemAssociationID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Verify if the association exists
	// GET /api/v2/users/{user_id}/systems
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/users/%s/systems", userID), nil)
	if err != nil {
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
		d.SetId("")
		return diag.Diagnostics{}
	}

	if err := d.Set("user_id", userID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("system_id", systemID); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

// resourceSystemDelete removes an association between a user and a system
func resourceSystemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Removing user-system association in JumpCloud: %s", d.Id()))

	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	// Extract user_id and system_id from the association ID
	userID, systemID, err := parseSystemAssociationID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// In JumpCloud, the API to remove an association is:
	// DELETE /api/v2/users/{user_id}/systems/{system_id}
	_, err = client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/users/%s/systems/%s", userID, systemID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error removing user-system association: %v", err))
	}

	// Clear the resource ID
	d.SetId("")

	return diag.Diagnostics{}
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
