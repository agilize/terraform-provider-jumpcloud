package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"username": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"username", "email", "id"},
			},
			"email": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"username", "email", "id"},
			},
			"user_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"username", "email", "id"},
			},
			"firstname": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"lastname": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attributes": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"mfa_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"password_never_expires": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	var path string
	var queryParam string
	var queryValue string

	if username, ok := d.GetOk("username"); ok {
		queryParam = "username"
		queryValue = username.(string)
	} else if email, ok := d.GetOk("email"); ok {
		queryParam = "email"
		queryValue = email.(string)
	} else if userID, ok := d.GetOk("user_id"); ok {
		path = fmt.Sprintf("/api/systemusers/%s", userID.(string))
	} else {
		return diag.FromErr(fmt.Errorf("one of username, email, or user_id must be provided"))
	}

	// If we're searching by username or email, we need to query the search endpoint
	if path == "" {
		path = fmt.Sprintf("/api/search/systemusers?%s=%s", queryParam, queryValue)
	}

	tflog.Debug(ctx, "Reading JumpCloud user data source", map[string]interface{}{
		"path": path,
	})

	// Usar a interface DoRequest em vez de acessar o cliente diretamente
	resp, err := c.DoRequest(http.MethodGet, path, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading user: %v", err))
	}

	var user User
	if err := json.Unmarshal(resp, &user); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %v", err))
	}

	d.SetId(user.ID)

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
