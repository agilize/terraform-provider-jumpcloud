package password_manager

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// DataSourceSafes returns a schema resource for JumpCloud password safes data source
func DataSourceSafes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSafesRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by safe name",
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by safe type (personal, team, shared)",
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by safe status (active, inactive)",
			},
			"owner_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by owner ID",
			},
			"member_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter safes that have a specific member",
			},
			"group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter safes that have a specific group",
			},
			"search": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter safes by text in name or description",
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     100,
				Description: "Maximum number of safes to return",
			},
			"skip": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Number of safes to skip",
			},
			"sort": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "name",
				Description: "Field to sort results by",
			},
			"sort_dir": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "asc",
				Description: "Sort direction (asc or desc)",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"safes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of password safes found",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the password safe",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the password safe",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Description of the password safe",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of safe (personal, team, shared)",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of the safe (active, inactive)",
						},
						"owner_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the user who owns the safe",
						},
						"member_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "IDs of users with access to the safe",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"group_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "IDs of user groups with access to the safe",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Organization ID",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation date of the safe",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Last update date of the safe",
						},
					},
				},
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of safes found",
			},
		},
	}
}

func dataSourceSafesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Build query parameters
	queryParams := constructSafesQueryParams(d)

	// Build URL with parameters
	url := fmt.Sprintf("/api/v2/password-safes?%s", queryParams)

	// Get safes via API
	tflog.Debug(ctx, fmt.Sprintf("Listing password safes with parameters: %s", queryParams))
	resp, err := client.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing password safes: %v", err))
	}

	// Deserialize response
	var safesResp SafesResponse
	if err := json.Unmarshal(resp, &safesResp); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Convert safes to Terraform format
	tfSafes := flattenSafes(safesResp.Results)
	if err := d.Set("safes", tfSafes); err != nil {
		return diag.FromErr(fmt.Errorf("error setting safes: %v", err))
	}

	d.Set("total", safesResp.TotalCount)

	// Generate unique ID for the data source
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

// Helper function to build query parameters
func constructSafesQueryParams(d *schema.ResourceData) string {
	params := ""

	// Add filters
	if v, ok := d.GetOk("name"); ok {
		params += fmt.Sprintf("name=%s&", v.(string))
	}

	if v, ok := d.GetOk("type"); ok {
		params += fmt.Sprintf("type=%s&", v.(string))
	}

	if v, ok := d.GetOk("status"); ok {
		params += fmt.Sprintf("status=%s&", v.(string))
	}

	if v, ok := d.GetOk("owner_id"); ok {
		params += fmt.Sprintf("owner_id=%s&", v.(string))
	}

	if v, ok := d.GetOk("member_id"); ok {
		params += fmt.Sprintf("member_id=%s&", v.(string))
	}

	if v, ok := d.GetOk("group_id"); ok {
		params += fmt.Sprintf("group_id=%s&", v.(string))
	}

	if v, ok := d.GetOk("search"); ok {
		params += fmt.Sprintf("search=%s&", v.(string))
	}

	// Add pagination and sorting
	if v, ok := d.GetOk("limit"); ok {
		params += fmt.Sprintf("limit=%d&", v.(int))
	}

	if v, ok := d.GetOk("skip"); ok {
		params += fmt.Sprintf("skip=%d&", v.(int))
	}

	if v, ok := d.GetOk("sort"); ok {
		params += fmt.Sprintf("sort=%s&", v.(string))
	}

	if v, ok := d.GetOk("sort_dir"); ok {
		params += fmt.Sprintf("sort_dir=%s&", v.(string))
	}

	if v, ok := d.GetOk("org_id"); ok {
		params += fmt.Sprintf("org_id=%s&", v.(string))
	}

	// Remove trailing & if present
	if len(params) > 0 && params[len(params)-1] == '&' {
		params = params[:len(params)-1]
	}

	return params
}

// Helper function to convert API response to Terraform format
func flattenSafes(safes []SafeItem) []map[string]interface{} {
	if safes == nil {
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, len(safes))
	for i, safe := range safes {
		safeMap := map[string]interface{}{
			"id":          safe.ID,
			"name":        safe.Name,
			"description": safe.Description,
			"type":        safe.Type,
			"status":      safe.Status,
			"owner_id":    safe.OwnerID,
			"created":     safe.Created,
			"updated":     safe.Updated,
		}

		if safe.MemberIDs != nil {
			safeMap["member_ids"] = safe.MemberIDs
		}

		if safe.GroupIDs != nil {
			safeMap["group_ids"] = safe.GroupIDs
		}

		if safe.OrgID != "" {
			safeMap["org_id"] = safe.OrgID
		}

		result[i] = safeMap
	}

	return result
}
