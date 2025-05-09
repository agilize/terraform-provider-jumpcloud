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

// ResourceAuthorization returns the schema for managing OAuth authorizations
func ResourceAuthorization() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAuthorizationCreate,
		ReadContext:   resourceAuthorizationRead,
		UpdateContext: resourceAuthorizationUpdate,
		DeleteContext: resourceAuthorizationDelete,
		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "OAuth application ID",
			},
			"expires_at": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Authorization expiration date (RFC3339 format)",
			},
			"client_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "OAuth client name",
			},
			"client_description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "OAuth client description",
			},
			"client_contact_email": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "OAuth client contact email",
			},
			"client_redirect_uris": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "OAuth client redirect URIs",
			},
			"scopes": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Scopes to be granted in the authorization",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Authorization creation date (RFC3339 format)",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update date (RFC3339 format)",
			},
			"org_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Organization ID",
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

// resourceAuthorizationCreate creates a new OAuth authorization
func resourceAuthorizationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Extract values from the schema
	applicationID := d.Get("application_id").(string)
	expiresAtStr := d.Get("expires_at").(string)

	// Convert expiresAt to time.Time
	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid format for expires_at: %v", err))
	}

	// Convert scopes to []string
	scopesRaw := d.Get("scopes").([]interface{})
	scopes := make([]string, len(scopesRaw))
	for i, v := range scopesRaw {
		scopes[i] = v.(string)
	}

	// Build object for API
	auth := &Authorization{
		ApplicationID: applicationID,
		ExpiresAt:     expiresAt,
		Scopes:        scopes,
	}

	// Add optional fields
	if v, ok := d.GetOk("client_name"); ok {
		auth.ClientName = v.(string)
	}
	if v, ok := d.GetOk("client_description"); ok {
		auth.ClientDescription = v.(string)
	}
	if v, ok := d.GetOk("client_contact_email"); ok {
		auth.ClientContactEmail = v.(string)
	}
	if v, ok := d.GetOk("client_redirect_uris"); ok {
		redirectURIsRaw := v.([]interface{})
		redirectURIs := make([]string, len(redirectURIsRaw))
		for i, uri := range redirectURIsRaw {
			redirectURIs[i] = uri.(string)
		}
		auth.ClientRedirectURIs = redirectURIs
	}

	// Serialize to JSON
	reqJSON, err := json.Marshal(auth)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing OAuth authorization: %v", err))
	}

	// Create OAuth authorization via API
	tflog.Debug(ctx, "Creating OAuth authorization", map[string]interface{}{
		"applicationId": applicationID,
	})

	resp, err := client.DoRequest(http.MethodPost, "/api/v2/oauth/authorizations", reqJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating OAuth authorization: %v", err))
	}

	// Deserialize response
	var createdAuth Authorization
	if err := json.Unmarshal(resp, &createdAuth); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set resource ID
	d.SetId(createdAuth.ID)

	// Read the resource to load all computed fields
	return resourceAuthorizationRead(ctx, d, meta)
}

// resourceAuthorizationRead reads an OAuth authorization
func resourceAuthorizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get resource ID
	id := d.Id()

	// Get OAuth authorization via API
	tflog.Debug(ctx, "Reading OAuth authorization", map[string]interface{}{
		"id": id,
	})

	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/oauth/authorizations/%s", id), nil)
	if err != nil {
		// Check if the resource no longer exists
		if common.IsNotFoundError(err) {
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("error reading OAuth authorization: %v", err))
	}

	// Deserialize response
	var auth Authorization
	if err := json.Unmarshal(resp, &auth); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Update state
	if err := d.Set("application_id", auth.ApplicationID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting application_id: %v", err))
	}

	if err := d.Set("expires_at", auth.ExpiresAt.Format(time.RFC3339)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting expires_at: %v", err))
	}

	if err := d.Set("client_name", auth.ClientName); err != nil {
		return diag.FromErr(fmt.Errorf("error setting client_name: %v", err))
	}

	if err := d.Set("client_description", auth.ClientDescription); err != nil {
		return diag.FromErr(fmt.Errorf("error setting client_description: %v", err))
	}

	if err := d.Set("client_contact_email", auth.ClientContactEmail); err != nil {
		return diag.FromErr(fmt.Errorf("error setting client_contact_email: %v", err))
	}

	if err := d.Set("client_redirect_uris", auth.ClientRedirectURIs); err != nil {
		return diag.FromErr(fmt.Errorf("error setting client_redirect_uris: %v", err))
	}

	if err := d.Set("scopes", auth.Scopes); err != nil {
		return diag.FromErr(fmt.Errorf("error setting scopes: %v", err))
	}

	if err := d.Set("org_id", auth.OrgID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting org_id: %v", err))
	}

	if !auth.Created.IsZero() {
		if err := d.Set("created", auth.Created.Format(time.RFC3339)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting created: %v", err))
		}
	}

	if !auth.Updated.IsZero() {
		if err := d.Set("updated", auth.Updated.Format(time.RFC3339)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting updated: %v", err))
		}
	}

	return diag.Diagnostics{}
}

// resourceAuthorizationUpdate updates an OAuth authorization
func resourceAuthorizationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get resource ID
	id := d.Id()

	// Check if any relevant field has changed
	if !d.HasChanges("expires_at", "client_name", "client_description", "client_contact_email", "client_redirect_uris", "scopes") {
		// No changes to be made
		return resourceAuthorizationRead(ctx, d, meta)
	}

	// Extract values from the schema
	applicationID := d.Get("application_id").(string)
	expiresAtStr := d.Get("expires_at").(string)

	// Convert expiresAt to time.Time
	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid format for expires_at: %v", err))
	}

	// Convert scopes to []string
	scopesRaw := d.Get("scopes").([]interface{})
	scopes := make([]string, len(scopesRaw))
	for i, v := range scopesRaw {
		scopes[i] = v.(string)
	}

	// Build object for API
	auth := &Authorization{
		ID:            id,
		ApplicationID: applicationID,
		ExpiresAt:     expiresAt,
		Scopes:        scopes,
	}

	// Add optional fields
	if v, ok := d.GetOk("client_name"); ok {
		auth.ClientName = v.(string)
	}
	if v, ok := d.GetOk("client_description"); ok {
		auth.ClientDescription = v.(string)
	}
	if v, ok := d.GetOk("client_contact_email"); ok {
		auth.ClientContactEmail = v.(string)
	}
	if v, ok := d.GetOk("client_redirect_uris"); ok {
		redirectURIsRaw := v.([]interface{})
		redirectURIs := make([]string, len(redirectURIsRaw))
		for i, uri := range redirectURIsRaw {
			redirectURIs[i] = uri.(string)
		}
		auth.ClientRedirectURIs = redirectURIs
	}

	// Serialize to JSON
	reqJSON, err := json.Marshal(auth)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing OAuth authorization: %v", err))
	}

	// Update OAuth authorization via API
	tflog.Debug(ctx, "Updating OAuth authorization", map[string]interface{}{
		"id": id,
	})

	_, err = client.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/oauth/authorizations/%s", id), reqJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating OAuth authorization: %v", err))
	}

	// Read the resource to reflect the changes
	return resourceAuthorizationRead(ctx, d, meta)
}

// resourceAuthorizationDelete deletes an OAuth authorization
func resourceAuthorizationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get resource ID
	id := d.Id()

	// Delete OAuth authorization via API
	tflog.Debug(ctx, "Deleting OAuth authorization", map[string]interface{}{
		"id": id,
	})

	_, err = client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/oauth/authorizations/%s", id), nil)
	if err != nil {
		// If the resource is already gone, just log
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, "OAuth authorization no longer exists", map[string]interface{}{
				"id": id,
			})
		} else {
			return diag.FromErr(fmt.Errorf("error deleting OAuth authorization: %v", err))
		}
	}

	// Remove ID from state
	d.SetId("")

	return diag.Diagnostics{}
}
