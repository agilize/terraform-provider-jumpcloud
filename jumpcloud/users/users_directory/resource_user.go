package users_directory

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// User represents a JumpCloud user
type User struct {
	ID                   string         `json:"_id,omitempty"`
	Username             string         `json:"username"`
	Email                string         `json:"email"`
	FirstName            string         `json:"firstname,omitempty"`
	LastName             string         `json:"lastname,omitempty"`
	Password             string         `json:"password,omitempty"`
	Description          string         `json:"description,omitempty"`
	Attributes           map[string]any `json:"attributes,omitempty"`
	MFAEnabled           bool           `json:"mfa_enabled,omitempty"`
	PasswordNeverExpires bool           `json:"password_never_expires,omitempty"`
}

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"firstname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"lastname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Custom attributes for the user (key-value pairs)",
			},
			"mfa_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"password_never_expires": {
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

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Build user object from resource data
	user := &User{
		Username:             d.Get("username").(string),
		Email:                d.Get("email").(string),
		FirstName:            d.Get("firstname").(string),
		LastName:             d.Get("lastname").(string),
		Password:             d.Get("password").(string),
		Description:          d.Get("description").(string),
		MFAEnabled:           d.Get("mfa_enabled").(bool),
		PasswordNeverExpires: d.Get("password_never_expires").(bool),
	}

	// Set custom attributes if present
	if v, ok := d.GetOk("attributes"); ok {
		user.Attributes = common.ExpandAttributes(v.(map[string]any))
	}

	// Convert user to JSON
	userJSON, err := json.Marshal(user)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing user: %v", err))
	}

	// Create user via API
	// The correct endpoint for creating users is /systemusers
	resp, err := c.DoRequest(http.MethodPost, "/systemusers", userJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating user: %v", err))
	}

	// Decode response
	var newUser User
	if err := json.Unmarshal(resp, &newUser); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing user response: %v", err))
	}

	// Set ID in resource data
	d.SetId(newUser.ID)

	// Read the user to set all the computed fields
	return resourceUserRead(ctx, d, meta)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	userID := d.Id()

	// Get user via API
	// The correct endpoint for getting users is /systemusers/{id}
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/systemusers/%s", userID), nil)
	if err != nil {
		// Handle 404 specifically to mark the resource as removed
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("User %s not found, removing from state", userID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading user %s: %v", userID, err))
	}

	// Decode response
	var user User
	if err := json.Unmarshal(resp, &user); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing user response: %v", err))
	}

	// Set fields in resource data
	d.Set("username", user.Username)
	d.Set("email", user.Email)
	d.Set("firstname", user.FirstName)
	d.Set("lastname", user.LastName)
	d.Set("description", user.Description)
	d.Set("mfa_enabled", user.MFAEnabled)
	d.Set("password_never_expires", user.PasswordNeverExpires)

	// Set custom attributes if present
	if len(user.Attributes) > 0 {
		d.Set("attributes", common.FlattenAttributes(user.Attributes))
	}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	userID := d.Id()

	// Build user object from resource data
	user := &User{
		Email:                d.Get("email").(string),
		FirstName:            d.Get("firstname").(string),
		LastName:             d.Get("lastname").(string),
		Description:          d.Get("description").(string),
		MFAEnabled:           d.Get("mfa_enabled").(bool),
		PasswordNeverExpires: d.Get("password_never_expires").(bool),
	}

	// Only set password if it's been changed
	if d.HasChange("password") {
		user.Password = d.Get("password").(string)
	}

	// Set custom attributes if present
	if v, ok := d.GetOk("attributes"); ok {
		user.Attributes = common.ExpandAttributes(v.(map[string]any))
	}

	// Convert user to JSON
	userJSON, err := json.Marshal(user)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing user: %v", err))
	}

	// Update user via API
	// The correct endpoint for updating users is /systemusers/{id}
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/systemusers/%s", userID), userJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating user %s: %v", userID, err))
	}

	return resourceUserRead(ctx, d, meta)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	userID := d.Id()

	// Delete user via API
	// The correct endpoint for deleting users is /systemusers/{id}
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/systemusers/%s", userID), nil)
	if err != nil {
		// If the resource is already gone, don't return an error
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("User %s not found during delete, removing from state", userID))
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting user %s: %v", userID, err))
	}

	// Set ID to empty to signify resource has been removed
	d.SetId("")

	return nil
}
