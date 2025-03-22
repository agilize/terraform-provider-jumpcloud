package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// AdminUsersResponse represents the API response for listing admin users
type AdminUsersResponse struct {
	Results    []AdminUser `json:"results"`
	TotalCount int         `json:"totalCount"`
}

// DataSourceUsers returns a schema.Resource for JumpCloud admin users
func DataSourceUsers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUsersRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"operator": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "eq",
							ValidateFunc: validation.StringInSlice([]string{"eq", "ne", "contains"}, false),
						},
					},
				},
				Description: "Filter criteria for the admin users",
			},
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 1000),
				Description:  "Maximum number of admin users to return",
			},
			"skip": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Number of admin users to skip",
			},
			"sort": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:     schema.TypeString,
							Required: true,
						},
						"direction": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"asc", "desc"}, false),
						},
					},
				},
				Description: "Sort criteria for the admin users",
			},
			"users": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"firstname": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"lastname": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"is_super_admin": {
							Type:     schema.TypeBool,
							Computed: true,
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
				},
				Description: "List of JumpCloud admin users",
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of admin users found",
			},
		},
	}
}

// dataSourceUsersRead reads admin users from JumpCloud
func dataSourceUsersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(JumpCloudClient)

	// Build query parameters
	queryParams := url.Values{}

	// Add filters if specified
	if filters, ok := d.GetOk("filter"); ok {
		for i, filter := range filters.([]interface{}) {
			f := filter.(map[string]interface{})
			name := f["name"].(string)
			value := f["value"].(string)
			operator := f["operator"].(string)

			queryParams.Add(fmt.Sprintf("filter[%d].field", i), name)
			queryParams.Add(fmt.Sprintf("filter[%d].value", i), value)
			queryParams.Add(fmt.Sprintf("filter[%d].operator", i), operator)
		}
	}

	// Add sort if specified
	if sortList, ok := d.GetOk("sort"); ok && len(sortList.([]interface{})) > 0 {
		sort := sortList.([]interface{})[0].(map[string]interface{})
		field := sort["field"].(string)
		direction := sort["direction"].(string)

		queryParams.Add("sort", field)
		queryParams.Add("sortDirection", direction)
	}

	// Add limit and skip
	queryParams.Add("limit", strconv.Itoa(d.Get("limit").(int)))
	queryParams.Add("skip", strconv.Itoa(d.Get("skip").(int)))

	// Construct URL
	urlStr := "/api/v2/administrators"
	if len(queryParams) > 0 {
		urlStr = fmt.Sprintf("%s?%s", urlStr, queryParams.Encode())
	}

	// Make the API request
	resp, err := client.DoRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading admin users: %v", err))
	}

	// Parse response
	var usersResp AdminUsersResponse
	if err := json.Unmarshal(resp, &usersResp); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing admin users response: %v", err))
	}

	// Set ID for the data source
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	// Flatten the users for the Terraform state
	users := flattenAdminUsers(usersResp.Results)

	// Set the users in the state
	if err := d.Set("users", users); err != nil {
		return diag.FromErr(fmt.Errorf("error setting users: %v", err))
	}

	// Set the total count
	if err := d.Set("total", usersResp.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("error setting total count: %v", err))
	}

	return nil
}

// flattenAdminUsers converts API user objects to a format suitable for Terraform state
func flattenAdminUsers(users []AdminUser) []map[string]interface{} {
	var result []map[string]interface{}

	for _, user := range users {
		userMap := map[string]interface{}{
			"id":             user.ID,
			"email":          user.Email,
			"firstname":      user.Firstname,
			"lastname":       user.Lastname,
			"is_super_admin": user.IsSuperAdmin,
			"created":        user.Created,
			"updated":        user.Updated,
		}

		result = append(result, userMap)
	}

	return result
}
