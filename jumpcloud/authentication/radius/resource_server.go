package radius

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// RadiusServer represents a RADIUS server in JumpCloud
type RadiusServer struct {
	ID                           string   `json:"_id,omitempty"`
	Name                         string   `json:"name"`
	SharedSecret                 string   `json:"sharedSecret"`
	NetworkSourceIP              string   `json:"networkSourceIp,omitempty"`
	MfaRequired                  bool     `json:"mfaRequired"`
	UserPasswordExpirationAction string   `json:"userPasswordExpirationAction,omitempty"`
	UserLockoutAction            string   `json:"userLockoutAction,omitempty"`
	UserAttribute                string   `json:"userAttribute,omitempty"`
	Targets                      []string `json:"targets,omitempty"`
	Created                      string   `json:"created,omitempty"`
	Updated                      string   `json:"updated,omitempty"`
}

// ResourceServer returns the resource for managing RADIUS servers
func ResourceServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerCreate,
		ReadContext:   resourceServerRead,
		UpdateContext: resourceServerUpdate,
		DeleteContext: resourceServerDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the RADIUS server",
			},
			"shared_secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Shared secret used for authentication between the client and RADIUS server",
			},
			"network_source_ip": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Network source IP that will be used to communicate with the RADIUS server",
			},
			"mfa_required": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether multi-factor authentication is required for the RADIUS server",
			},
			"user_password_expiration_action": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "allow",
				ValidateFunc: validation.StringInSlice([]string{"allow", "deny"}, false),
				Description:  "Action to take when a user's password expires (allow or deny)",
			},
			"user_lockout_action": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "deny",
				ValidateFunc: validation.StringInSlice([]string{"allow", "deny"}, false),
				Description:  "Action to take when a user is locked out (allow or deny)",
			},
			"user_attribute": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "username",
				ValidateFunc: validation.StringInSlice([]string{"username", "email"}, false),
				Description:  "User attribute used for authentication (username or email)",
			},
			"targets": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of group IDs associated with the RADIUS server",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the RADIUS server",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update date of the RADIUS server",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manages RADIUS servers in JumpCloud. This resource allows creating, updating, and deleting RADIUS server configurations.",
	}
}

// resourceServerCreate creates a new RADIUS server in JumpCloud
func resourceServerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client
	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	// Build RADIUS server
	radiusServer := &RadiusServer{
		Name:         d.Get("name").(string),
		SharedSecret: d.Get("shared_secret").(string),
		MfaRequired:  d.Get("mfa_required").(bool),
	}

	// Optional fields
	if v, ok := d.GetOk("network_source_ip"); ok {
		radiusServer.NetworkSourceIP = v.(string)
	}

	if v, ok := d.GetOk("user_password_expiration_action"); ok {
		radiusServer.UserPasswordExpirationAction = v.(string)
	}

	if v, ok := d.GetOk("user_lockout_action"); ok {
		radiusServer.UserLockoutAction = v.(string)
	}

	if v, ok := d.GetOk("user_attribute"); ok {
		radiusServer.UserAttribute = v.(string)
	}

	if v, ok := d.GetOk("targets"); ok {
		targets := v.([]interface{})
		radiusServer.Targets = make([]string, len(targets))
		for i, target := range targets {
			radiusServer.Targets[i] = target.(string)
		}
	}

	// Serialize to JSON
	radiusServerJSON, err := json.Marshal(radiusServer)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing RADIUS server: %v", err))
	}

	// Create RADIUS server via API
	tflog.Debug(ctx, fmt.Sprintf("Creating RADIUS server: %s", radiusServer.Name))
	resp, err := client.DoRequest(http.MethodPost, "/api/v2/radiusservers", radiusServerJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating RADIUS server: %v", err))
	}

	// Deserialize response
	var createdRadiusServer RadiusServer
	if err := json.Unmarshal(resp, &createdRadiusServer); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	if createdRadiusServer.ID == "" {
		return diag.FromErr(fmt.Errorf("RADIUS server created without ID"))
	}

	d.SetId(createdRadiusServer.ID)
	return resourceServerRead(ctx, d, meta)
}

// resourceServerRead reads the details of a RADIUS server from JumpCloud
func resourceServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get client
	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("RADIUS server ID not provided"))
	}

	// Fetch RADIUS server via API
	tflog.Debug(ctx, fmt.Sprintf("Reading RADIUS server with ID: %s", id))
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/radiusservers/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("RADIUS server %s not found, removing from state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading RADIUS server: %v", err))
	}

	// Deserialize response
	var radiusServer RadiusServer
	if err := json.Unmarshal(resp, &radiusServer); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set values in state
	if err := d.Set("name", radiusServer.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %v", err))
	}

	// We don't set shared_secret in state to avoid exposing credentials
	if err := d.Set("network_source_ip", radiusServer.NetworkSourceIP); err != nil {
		return diag.FromErr(fmt.Errorf("error setting network_source_ip: %v", err))
	}

	if err := d.Set("mfa_required", radiusServer.MfaRequired); err != nil {
		return diag.FromErr(fmt.Errorf("error setting mfa_required: %v", err))
	}

	if err := d.Set("user_password_expiration_action", radiusServer.UserPasswordExpirationAction); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_password_expiration_action: %v", err))
	}

	if err := d.Set("user_lockout_action", radiusServer.UserLockoutAction); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_lockout_action: %v", err))
	}

	if err := d.Set("user_attribute", radiusServer.UserAttribute); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_attribute: %v", err))
	}

	if err := d.Set("created", radiusServer.Created); err != nil {
		return diag.FromErr(fmt.Errorf("error setting created: %v", err))
	}

	if err := d.Set("updated", radiusServer.Updated); err != nil {
		return diag.FromErr(fmt.Errorf("error setting updated: %v", err))
	}

	if radiusServer.Targets != nil {
		if err := d.Set("targets", radiusServer.Targets); err != nil {
			return diag.FromErr(fmt.Errorf("error setting targets: %v", err))
		}
	}

	return diags
}

// resourceServerUpdate updates an existing RADIUS server in JumpCloud
func resourceServerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client
	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("RADIUS server ID not provided"))
	}

	// Build updated RADIUS server
	radiusServer := &RadiusServer{
		ID:          id,
		Name:        d.Get("name").(string),
		MfaRequired: d.Get("mfa_required").(bool),
	}

	// Always include shared secret for updates
	radiusServer.SharedSecret = d.Get("shared_secret").(string)

	// Optional fields
	if v, ok := d.GetOk("network_source_ip"); ok {
		radiusServer.NetworkSourceIP = v.(string)
	}

	if v, ok := d.GetOk("user_password_expiration_action"); ok {
		radiusServer.UserPasswordExpirationAction = v.(string)
	}

	if v, ok := d.GetOk("user_lockout_action"); ok {
		radiusServer.UserLockoutAction = v.(string)
	}

	if v, ok := d.GetOk("user_attribute"); ok {
		radiusServer.UserAttribute = v.(string)
	}

	if v, ok := d.GetOk("targets"); ok {
		targets := v.([]interface{})
		radiusServer.Targets = make([]string, len(targets))
		for i, target := range targets {
			radiusServer.Targets[i] = target.(string)
		}
	}

	// Serialize to JSON
	radiusServerJSON, err := json.Marshal(radiusServer)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing RADIUS server: %v", err))
	}

	// Update RADIUS server via API
	tflog.Debug(ctx, fmt.Sprintf("Updating RADIUS server with ID: %s", id))
	resp, err := client.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/radiusservers/%s", id), radiusServerJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating RADIUS server: %v", err))
	}

	// Deserialize response
	var updatedRadiusServer RadiusServer
	if err := json.Unmarshal(resp, &updatedRadiusServer); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	return resourceServerRead(ctx, d, meta)
}

// resourceServerDelete deletes a RADIUS server from JumpCloud
func resourceServerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get client
	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("RADIUS server ID not provided"))
	}

	// Delete RADIUS server via API
	tflog.Debug(ctx, fmt.Sprintf("Deleting RADIUS server with ID: %s", id))
	_, err := client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/radiusservers/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			// If the server doesn't exist, consider the delete successful
			tflog.Warn(ctx, fmt.Sprintf("RADIUS server %s not found, assuming already deleted", id))
			return diags
		}
		return diag.FromErr(fmt.Errorf("error deleting RADIUS server %s: %v", id, err))
	}

	// Clear the ID
	d.SetId("")
	return diags
}
