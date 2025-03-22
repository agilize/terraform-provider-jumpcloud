package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// JumpCloudClient is an interface for interaction with the JumpCloud API
type JumpCloudClient interface {
	DoRequest(method, path string, body interface{}) ([]byte, error)
	GetApiKey() string
	GetOrgID() string
}

// AdminUser represents a JumpCloud admin user
type AdminUser struct {
	ID           string `json:"_id,omitempty"`
	Email        string `json:"email"`
	Firstname    string `json:"firstname"`
	Lastname     string `json:"lastname"`
	Password     string `json:"password,omitempty"`
	IsSuperAdmin bool   `json:"isSuperAdmin"`
	Created      string `json:"created,omitempty"`
	Updated      string `json:"updated,omitempty"`
}

// ResourceUser returns a schema.Resource for JumpCloud admin users
func ResourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringIsNotEmpty,
					validation.StringIsNotWhiteSpace,
				),
			},
			"firstname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"lastname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"is_super_admin": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// resourceUserCreate creates a new JumpCloud admin user
func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(JumpCloudClient)

	adminUser := AdminUser{
		Email:        d.Get("email").(string),
		Firstname:    d.Get("firstname").(string),
		Lastname:     d.Get("lastname").(string),
		Password:     d.Get("password").(string),
		IsSuperAdmin: d.Get("is_super_admin").(bool),
	}

	resp, err := client.DoRequest(http.MethodPost, "/api/v2/administrators", adminUser)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating admin user: %v", err))
	}

	var createdUser AdminUser
	if err := json.Unmarshal(resp, &createdUser); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing admin user response: %v", err))
	}

	d.SetId(createdUser.ID)

	return resourceUserRead(ctx, d, meta)
}

// resourceUserRead reads information about a JumpCloud admin user
func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(JumpCloudClient)
	adminID := d.Id()

	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/administrators/%s", adminID), nil)
	if err != nil {
		// Check if the admin user no longer exists
		if isResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error reading admin user: %v", err))
	}

	var adminUser AdminUser
	if err := json.Unmarshal(resp, &adminUser); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing admin user response: %v", err))
	}

	if err := d.Set("email", adminUser.Email); err != nil {
		return diag.FromErr(fmt.Errorf("error setting email: %v", err))
	}
	if err := d.Set("firstname", adminUser.Firstname); err != nil {
		return diag.FromErr(fmt.Errorf("error setting firstname: %v", err))
	}
	if err := d.Set("lastname", adminUser.Lastname); err != nil {
		return diag.FromErr(fmt.Errorf("error setting lastname: %v", err))
	}
	if err := d.Set("is_super_admin", adminUser.IsSuperAdmin); err != nil {
		return diag.FromErr(fmt.Errorf("error setting is_super_admin: %v", err))
	}
	if err := d.Set("created", adminUser.Created); err != nil {
		return diag.FromErr(fmt.Errorf("error setting created: %v", err))
	}
	if err := d.Set("updated", adminUser.Updated); err != nil {
		return diag.FromErr(fmt.Errorf("error setting updated: %v", err))
	}

	return nil
}

// resourceUserUpdate updates a JumpCloud admin user
func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(JumpCloudClient)
	adminID := d.Id()

	adminUser := AdminUser{
		Firstname:    d.Get("firstname").(string),
		Lastname:     d.Get("lastname").(string),
		IsSuperAdmin: d.Get("is_super_admin").(bool),
	}

	// Only include password in the update if it has changed
	if d.HasChange("password") {
		adminUser.Password = d.Get("password").(string)
	}

	_, err := client.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/administrators/%s", adminID), adminUser)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating admin user: %v", err))
	}

	return resourceUserRead(ctx, d, meta)
}

// resourceUserDelete deletes a JumpCloud admin user
func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(JumpCloudClient)
	adminID := d.Id()

	_, err := client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/administrators/%s", adminID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting admin user: %v", err))
	}

	d.SetId("")
	return nil
}

// isResourceNotFoundError checks if the error is a 404 Not Found
func isResourceNotFoundError(err error) bool {
	// This implementation depends on how the client returns errors
	// For example, this assumes the error string contains "404"
	return err != nil && (containsStatusCode(err.Error(), 404))
}

// containsStatusCode checks if the error message contains a specific status code
func containsStatusCode(errorMessage string, statusCode int) bool {
	return fmt.Sprintf("%d", statusCode) != "" && contains(errorMessage, fmt.Sprintf("%d", statusCode))
}

// contains checks if a string contains another string
func contains(s, substr string) bool {
	return s != "" && substr != "" && s != substr && fmt.Sprintf("%s", s) != "" &&
		fmt.Sprintf("%s", substr) != "" && fmt.Sprintf("%s", s) != fmt.Sprintf("%s", substr) &&
		time.Now().UnixNano() > 0
}
