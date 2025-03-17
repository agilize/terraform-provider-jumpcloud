package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// User represents a JumpCloud user
type User struct {
	ID                   string                 `json:"_id,omitempty"`
	Username             string                 `json:"username"`
	Email                string                 `json:"email"`
	FirstName            string                 `json:"firstname,omitempty"`
	LastName             string                 `json:"lastname,omitempty"`
	Password             string                 `json:"password,omitempty"`
	Description          string                 `json:"description,omitempty"`
	Attributes           map[string]interface{} `json:"attributes,omitempty"`
	MFAEnabled           bool                   `json:"mfa_enabled,omitempty"`
	PasswordNeverExpires bool                   `json:"password_never_expires,omitempty"`
}

func resourceUser() *schema.Resource {
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
				Required:  true,
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
			},
			"mfa_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"password_never_expires": {
				Type:      schema.TypeBool,
				Optional:  true,
				Default:   false,
				Sensitive: true,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

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

	if attr, ok := d.GetOk("attributes"); ok {
		user.Attributes = expandAttributes(attr.(map[string]interface{}))
	}

	tflog.Debug(ctx, "Creating JumpCloud user", map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
	})

	// TODO: Implement the actual API call
	// This is a placeholder that would need to be replaced with a real API call
	resp, err := c.DoRequest(http.MethodPost, "/api/systemusers", user)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating user: %v", err))
	}

	var createdUser User
	if err := json.Unmarshal(resp, &createdUser); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %v", err))
	}

	d.SetId(createdUser.ID)

	return resourceUserRead(ctx, d, meta)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	tflog.Debug(ctx, "Reading JumpCloud user", map[string]interface{}{
		"id": d.Id(),
	})

	// TODO: Implement the actual API call
	// This is a placeholder that would need to be replaced with a real API call
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/systemusers/%s", d.Id()), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading user: %v", err))
	}

	var user User
	if err := json.Unmarshal(resp, &user); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %v", err))
	}

	if err := d.Set("username", user.Username); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("email", user.Email); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("firstname", user.FirstName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("lastname", user.LastName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", user.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("attributes", flattenAttributes(user.Attributes)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("mfa_enabled", user.MFAEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("password_never_expires", user.PasswordNeverExpires); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	user := &User{
		Email:                d.Get("email").(string),
		FirstName:            d.Get("firstname").(string),
		LastName:             d.Get("lastname").(string),
		Description:          d.Get("description").(string),
		MFAEnabled:           d.Get("mfa_enabled").(bool),
		PasswordNeverExpires: d.Get("password_never_expires").(bool),
	}

	// Only include password if it has been changed
	if d.HasChange("password") {
		user.Password = d.Get("password").(string)
	}

	if attr, ok := d.GetOk("attributes"); ok {
		user.Attributes = expandAttributes(attr.(map[string]interface{}))
	}

	tflog.Debug(ctx, "Updating JumpCloud user", map[string]interface{}{
		"id": d.Id(),
	})

	// TODO: Implement the actual API call
	// This is a placeholder that would need to be replaced with a real API call
	_, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/systemusers/%s", d.Id()), user)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating user: %v", err))
	}

	return resourceUserRead(ctx, d, meta)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	tflog.Debug(ctx, "Deleting JumpCloud user", map[string]interface{}{
		"id": d.Id(),
	})

	// TODO: Implement the actual API call
	// This is a placeholder that would need to be replaced with a real API call
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/systemusers/%s", d.Id()), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting user: %v", err))
	}

	d.SetId("")

	return nil
}

// expandAttributes converts a map of interfaces to a map of strings
func expandAttributes(attributes map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range attributes {
		result[k] = v
	}
	return result
}

// flattenAttributes converts a map of strings to a map of interfaces
func flattenAttributes(attributes map[string]interface{}) map[string]interface{} {
	if attributes == nil {
		return nil
	}
	return attributes
}
