package scim

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// ScimServer represents a SCIM server in JumpCloud
type ScimServer struct {
	ID            string                 `json:"_id,omitempty"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	Type          string                 `json:"type"` // azure_ad, okta, generic, etc.
	BaseURL       string                 `json:"baseUrl,omitempty"`
	Enabled       bool                   `json:"enabled"`
	Status        string                 `json:"status,omitempty"` // active, error, etc.
	AuthType      string                 `json:"authType"`         // token, basic, oauth
	AuthConfig    map[string]interface{} `json:"authConfig"`
	CustomHeaders map[string]string      `json:"customHeaders,omitempty"`
	Features      []string               `json:"features,omitempty"` // users, groups, etc.
	Mappings      map[string]interface{} `json:"mappings,omitempty"`
	OrgID         string                 `json:"orgId,omitempty"`
	Created       string                 `json:"created,omitempty"`
	Updated       string                 `json:"updated,omitempty"`
}

// ResourceServer returns a schema resource for managing SCIM servers in JumpCloud
func ResourceServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerCreate,
		ReadContext:   resourceServerRead,
		UpdateContext: resourceServerUpdate,
		DeleteContext: resourceServerDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
				Description:  "Name of the SCIM server",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the SCIM server",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"azure_ad", "okta", "generic", "one_login", "google", "idp", "workspace",
				}, false),
				Description: "Type of the SCIM server (azure_ad, okta, generic, etc.)",
			},
			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Base URL for the SCIM server",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicates if the SCIM server is enabled",
			},
			"auth_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"token", "basic", "oauth",
				}, false),
				Description: "Authentication type (token, basic, oauth)",
			},
			"auth_config": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: common.SuppressEquivalentJSONDiffs,
				Description:      "Authentication configuration in JSON format (sensitive)",
				Sensitive:        true,
			},
			"custom_headers": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Custom HTTP headers",
			},
			"features": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"users", "groups", "provisioning",
					}, false),
				},
				Description: "Enabled features (users, groups, provisioning)",
			},
			"mappings": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: common.SuppressEquivalentJSONDiffs,
				Description:      "Attribute mappings in JSON format",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current status of the SCIM server",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the server",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update date of the server",
			},
		},
	}
}

func resourceServerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetJumpCloudClient(meta)
	if diagErr != nil {
		return diagErr
	}

	// Build ScimServer object from terraform data
	server := &ScimServer{
		Name:     d.Get("name").(string),
		Type:     d.Get("type").(string),
		Enabled:  d.Get("enabled").(bool),
		AuthType: d.Get("auth_type").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		server.Description = v.(string)
	}

	if v, ok := d.GetOk("base_url"); ok {
		server.BaseURL = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		server.OrgID = v.(string)
	}

	// Process authentication configuration (JSON)
	authConfigJSON := d.Get("auth_config").(string)
	var authConfig map[string]interface{}
	if err := json.Unmarshal([]byte(authConfigJSON), &authConfig); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing auth_config: %v", err))
	}
	server.AuthConfig = authConfig

	// Process custom headers
	if v, ok := d.GetOk("custom_headers"); ok {
		customHeaders := make(map[string]string)
		for k, v := range v.(map[string]interface{}) {
			customHeaders[k] = v.(string)
		}
		server.CustomHeaders = customHeaders
	}

	// Process features
	if v, ok := d.GetOk("features"); ok {
		featuresSet := v.(*schema.Set)
		features := make([]string, featuresSet.Len())
		for i, f := range featuresSet.List() {
			features[i] = f.(string)
		}
		server.Features = features
	}

	// Process mappings (JSON)
	if v, ok := d.GetOk("mappings"); ok {
		mappingsJSON := v.(string)
		var mappings map[string]interface{}
		if err := json.Unmarshal([]byte(mappingsJSON), &mappings); err != nil {
			return diag.FromErr(fmt.Errorf("error deserializing mappings: %v", err))
		}
		server.Mappings = mappings
	}

	// Serialize to JSON
	reqBody, err := json.Marshal(server)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing SCIM server: %v", err))
	}

	// Build URL for request
	url := "/api/v2/scim/servers"
	if server.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, server.OrgID)
	}

	// Make request to create server
	tflog.Debug(ctx, "Creating SCIM server")
	resp, err := c.DoRequest(http.MethodPost, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating SCIM server: %v", err))
	}

	// Deserialize response
	var createdServer ScimServer
	if err := json.Unmarshal(resp, &createdServer); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	if createdServer.ID == "" {
		return diag.FromErr(fmt.Errorf("created SCIM server returned without an ID"))
	}

	// Set ID in state
	d.SetId(createdServer.ID)

	// Read the resource to update state with all computed fields
	return resourceServerRead(ctx, d, meta)
}

func resourceServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetJumpCloudClient(meta)
	if diagErr != nil {
		return diagErr
	}

	// Get server ID
	serverID := d.Id()
	if serverID == "" {
		return diag.FromErr(fmt.Errorf("SCIM server ID is required"))
	}

	// Get orgId parameter if available
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Make request to get server details
	tflog.Debug(ctx, fmt.Sprintf("Reading SCIM server: %s", serverID))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/scim/servers/%s%s", serverID, orgIDParam), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("SCIM server %s not found, removing from state", serverID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading SCIM server: %v", err))
	}

	// Deserialize response
	var server ScimServer
	if err := json.Unmarshal(resp, &server); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing SCIM server: %v", err))
	}

	// Set values in state
	if err := d.Set("name", server.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %v", err))
	}

	if err := d.Set("description", server.Description); err != nil {
		return diag.FromErr(fmt.Errorf("error setting description: %v", err))
	}

	if err := d.Set("type", server.Type); err != nil {
		return diag.FromErr(fmt.Errorf("error setting type: %v", err))
	}

	if err := d.Set("base_url", server.BaseURL); err != nil {
		return diag.FromErr(fmt.Errorf("error setting base_url: %v", err))
	}

	if err := d.Set("enabled", server.Enabled); err != nil {
		return diag.FromErr(fmt.Errorf("error setting enabled: %v", err))
	}

	if err := d.Set("auth_type", server.AuthType); err != nil {
		return diag.FromErr(fmt.Errorf("error setting auth_type: %v", err))
	}

	if err := d.Set("status", server.Status); err != nil {
		return diag.FromErr(fmt.Errorf("error setting status: %v", err))
	}

	if err := d.Set("created", server.Created); err != nil {
		return diag.FromErr(fmt.Errorf("error setting created: %v", err))
	}

	if err := d.Set("updated", server.Updated); err != nil {
		return diag.FromErr(fmt.Errorf("error setting updated: %v", err))
	}

	// Set custom headers
	if server.CustomHeaders != nil {
		if err := d.Set("custom_headers", server.CustomHeaders); err != nil {
			return diag.FromErr(fmt.Errorf("error setting custom_headers: %v", err))
		}
	}

	// Set features
	if server.Features != nil {
		if err := d.Set("features", server.Features); err != nil {
			return diag.FromErr(fmt.Errorf("error setting features: %v", err))
		}
	}

	// Set auth_config and mappings - Note: these require special handling
	// We don't update these fields from the API response to avoid exposing sensitive data
	// and to prevent unnecessary diffs due to JSON formatting differences

	return diags
}

func resourceServerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetJumpCloudClient(meta)
	if diagErr != nil {
		return diagErr
	}

	// Get server ID
	serverID := d.Id()
	if serverID == "" {
		return diag.FromErr(fmt.Errorf("SCIM server ID is required"))
	}

	// Build server object for update
	server := &ScimServer{
		ID:       serverID,
		Name:     d.Get("name").(string),
		Type:     d.Get("type").(string),
		Enabled:  d.Get("enabled").(bool),
		AuthType: d.Get("auth_type").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		server.Description = v.(string)
	}

	if v, ok := d.GetOk("base_url"); ok {
		server.BaseURL = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		server.OrgID = v.(string)
	}

	// Process authentication configuration (JSON)
	authConfigJSON := d.Get("auth_config").(string)
	var authConfig map[string]interface{}
	if err := json.Unmarshal([]byte(authConfigJSON), &authConfig); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing auth_config: %v", err))
	}
	server.AuthConfig = authConfig

	// Process custom headers
	if v, ok := d.GetOk("custom_headers"); ok {
		customHeaders := make(map[string]string)
		for k, v := range v.(map[string]interface{}) {
			customHeaders[k] = v.(string)
		}
		server.CustomHeaders = customHeaders
	}

	// Process features
	if v, ok := d.GetOk("features"); ok {
		featuresSet := v.(*schema.Set)
		features := make([]string, featuresSet.Len())
		for i, f := range featuresSet.List() {
			features[i] = f.(string)
		}
		server.Features = features
	}

	// Process mappings (JSON)
	if v, ok := d.GetOk("mappings"); ok {
		mappingsJSON := v.(string)
		var mappings map[string]interface{}
		if err := json.Unmarshal([]byte(mappingsJSON), &mappings); err != nil {
			return diag.FromErr(fmt.Errorf("error deserializing mappings: %v", err))
		}
		server.Mappings = mappings
	}

	// Serialize to JSON
	reqBody, err := json.Marshal(server)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing SCIM server: %v", err))
	}

	// Build URL for request
	url := fmt.Sprintf("/api/v2/scim/servers/%s", serverID)
	if server.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, server.OrgID)
	}

	// Make request to update server
	tflog.Debug(ctx, fmt.Sprintf("Updating SCIM server: %s", serverID))
	_, err = c.DoRequest(http.MethodPut, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating SCIM server: %v", err))
	}

	return resourceServerRead(ctx, d, meta)
}

func resourceServerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetJumpCloudClient(meta)
	if diagErr != nil {
		return diagErr
	}

	// Get server ID
	serverID := d.Id()
	if serverID == "" {
		return diag.FromErr(fmt.Errorf("SCIM server ID is required"))
	}

	// Get orgId parameter if available
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Make request to delete server
	tflog.Debug(ctx, fmt.Sprintf("Deleting SCIM server: %s", serverID))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/scim/servers/%s%s", serverID, orgIDParam), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting SCIM server: %v", err))
	}

	// Clear ID from state
	d.SetId("")

	return diags
}
