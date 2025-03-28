package oauth

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

// ResourceUser returns the schema for managing OAuth users
func ResourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "OAuth application ID",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "JumpCloud user ID",
			},
			"scopes": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "OAuth scopes to be granted to the user",
			},
			"username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Username (read-only)",
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "User's email (read-only)",
			},
			"first_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "User's first name (read-only)",
			},
			"last_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "User's last name (read-only)",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date (RFC3339 format)",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update date (RFC3339 format)",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Read:   schema.DefaultTimeout(30 * time.Second),
			Update: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

// resourceUserCreate creates a new OAuth user
func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Extract values from the schema
	applicationID := d.Get("application_id").(string)
	userID := d.Get("user_id").(string)
	scopesRaw := d.Get("scopes").([]interface{})

	// Convert scopes to []string
	scopes := make([]string, len(scopesRaw))
	for i, v := range scopesRaw {
		scopes[i] = v.(string)
	}

	// Build object for API
	oauthUser := &User{
		ApplicationID: applicationID,
		UserID:        userID,
		Scopes:        scopes,
	}

	// Serialize to JSON
	reqJSON, err := json.Marshal(oauthUser)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing OAuth user: %v", err))
	}

	// Create OAuth user via API
	tflog.Debug(ctx, "Creating OAuth user", map[string]interface{}{
		"applicationId": applicationID,
		"userId":        userID,
	})

	resp, err := client.DoRequest(http.MethodPost, "/api/v2/oauth/users", reqJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating OAuth user: %v", err))
	}

	// Deserialize response
	var createdUser User
	if err := json.Unmarshal(resp, &createdUser); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set resource ID
	d.SetId(createdUser.ID)

	// Read the resource to load all computed fields
	return resourceUserRead(ctx, d, meta)
}

// resourceUserRead reads an OAuth user
func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get resource ID
	id := d.Id()

	// Get OAuth user via API
	tflog.Debug(ctx, "Reading OAuth user", map[string]interface{}{
		"id": id,
	})

	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/oauth/users/%s", id), nil)
	if err != nil {
		// Check if the resource no longer exists
		if common.IsNotFoundError(err) {
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("error reading OAuth user: %v", err))
	}

	// Deserialize response
	var oauthUser User
	if err := json.Unmarshal(resp, &oauthUser); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Update state
	d.Set("application_id", oauthUser.ApplicationID)
	d.Set("user_id", oauthUser.UserID)
	d.Set("scopes", oauthUser.Scopes)
	d.Set("username", oauthUser.Username)
	d.Set("email", oauthUser.Email)
	d.Set("first_name", oauthUser.FirstName)
	d.Set("last_name", oauthUser.LastName)

	if !oauthUser.Created.IsZero() {
		d.Set("created", oauthUser.Created.Format(time.RFC3339))
	}
	if !oauthUser.Updated.IsZero() {
		d.Set("updated", oauthUser.Updated.Format(time.RFC3339))
	}

	return diag.Diagnostics{}
}

// resourceUserUpdate updates an OAuth user
func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get resource ID
	id := d.Id()

	// Check if any field was changed
	if !d.HasChange("scopes") {
		// No changes to be made
		return resourceUserRead(ctx, d, meta)
	}

	// Extract values from the schema
	applicationID := d.Get("application_id").(string)
	userID := d.Get("user_id").(string)
	scopesRaw := d.Get("scopes").([]interface{})

	// Convert scopes to []string
	scopes := make([]string, len(scopesRaw))
	for i, v := range scopesRaw {
		scopes[i] = v.(string)
	}

	// Build object for API
	oauthUser := &User{
		ID:            id,
		ApplicationID: applicationID,
		UserID:        userID,
		Scopes:        scopes,
	}

	// Serialize to JSON
	reqJSON, err := json.Marshal(oauthUser)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing OAuth user: %v", err))
	}

	// Update OAuth user via API
	tflog.Debug(ctx, "Updating OAuth user", map[string]interface{}{
		"id": id,
	})

	_, err = client.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/oauth/users/%s", id), reqJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating OAuth user: %v", err))
	}

	// Read the resource to reflect the changes
	return resourceUserRead(ctx, d, meta)
}

// resourceUserDelete deletes an OAuth user
func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get resource ID
	id := d.Id()

	// Delete OAuth user via API
	tflog.Debug(ctx, "Deleting OAuth user", map[string]interface{}{
		"id": id,
	})

	_, err = client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/oauth/users/%s", id), nil)
	if err != nil {
		// If the resource is already gone, just log
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, "OAuth user no longer exists", map[string]interface{}{
				"id": id,
			})
		} else {
			return diag.FromErr(fmt.Errorf("error deleting OAuth user: %v", err))
		}
	}

	// Remove ID from state
	d.SetId("")

	return diag.Diagnostics{}
}
