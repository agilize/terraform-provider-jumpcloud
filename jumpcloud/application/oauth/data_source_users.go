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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// DataSourceUsers returns a schema for the OAuth users data source
func DataSourceUsers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUsersRead,
		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "OAuth application ID to filter users",
			},
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 1000),
				Description:  "Maximum number of users to return (1-1000)",
			},
			"skip": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Number of users to skip (pagination)",
			},
			"sort": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "username",
				Description: "Field to sort results by",
			},
			"sort_dir": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "asc",
				ValidateFunc: validation.StringInSlice([]string{"asc", "desc"}, false),
				Description:  "Sort direction: asc (ascending) or desc (descending)",
			},
			"filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Expression to filter users (e.g. 'username:contains:john')",
			},
			"users": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Unique ID of the OAuth user record",
						},
						"application_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "OAuth application ID",
						},
						"user_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "JumpCloud user ID",
						},
						"username": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Username",
						},
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "User's email",
						},
						"first_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "User's first name",
						},
						"last_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "User's last name",
						},
						"scopes": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "OAuth scopes granted to the user",
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
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of OAuth users matching the criteria",
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether there are more users available beyond those returned",
			},
			"next_offset": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Offset for the next page of results",
			},
		},
	}
}

// dataSourceUsersRead reads OAuth users
func dataSourceUsersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Extract search parameters
	applicationID := d.Get("application_id").(string)
	limit := d.Get("limit").(int)
	skip := d.Get("skip").(int)
	sort := d.Get("sort").(string)
	sortDir := d.Get("sort_dir").(string)

	// Build request
	req := &UserRequest{
		ApplicationID: applicationID,
		Limit:         limit,
		Skip:          skip,
		Sort:          sort,
		SortDir:       sortDir,
	}

	// Add filter if provided
	if v, ok := d.GetOk("filter"); ok {
		filterStr := v.(string)
		req.Filter = &filterStr
	}

	// Serialize to JSON
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing request: %v", err))
	}

	// Fetch OAuth users via API
	tflog.Debug(ctx, "Fetching OAuth users", map[string]interface{}{
		"applicationId": applicationID,
		"limit":         limit,
		"skip":          skip,
	})

	resp, err := client.DoRequest(http.MethodPost, "/api/v2/oauth/users/search", reqJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching OAuth users: %v", err))
	}

	// Deserialize response
	var usersResp UsersResponse
	if err := json.Unmarshal(resp, &usersResp); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Process users and set in state
	users := make([]map[string]interface{}, len(usersResp.Results))
	for i, user := range usersResp.Results {
		users[i] = map[string]interface{}{
			"id":             user.ID,
			"application_id": user.ApplicationID,
			"user_id":        user.UserID,
			"username":       user.Username,
			"email":          user.Email,
			"first_name":     user.FirstName,
			"last_name":      user.LastName,
			"scopes":         user.Scopes,
			"created":        user.Created.Format(time.RFC3339),
			"updated":        user.Updated.Format(time.RFC3339),
		}
	}

	// Update state
	d.SetId(time.Now().Format(time.RFC3339)) // Unique ID for the data source

	if err := d.Set("users", users); err != nil {
		return diag.FromErr(fmt.Errorf("error setting users: %v", err))
	}

	if err := d.Set("total_count", usersResp.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("error setting total_count: %v", err))
	}

	if err := d.Set("has_more", usersResp.HasMore); err != nil {
		return diag.FromErr(fmt.Errorf("error setting has_more: %v", err))
	}

	if err := d.Set("next_offset", usersResp.NextOffset); err != nil {
		return diag.FromErr(fmt.Errorf("error setting next_offset: %v", err))
	}

	return diag.Diagnostics{}
}
