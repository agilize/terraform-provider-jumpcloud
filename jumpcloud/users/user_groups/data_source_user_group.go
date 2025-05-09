package users

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

// DataSourceUserGroup returns a schema for the JumpCloud user group data source
func DataSourceUserGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserGroupRead,
		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"name"},
				Description:   "The ID of the user group to retrieve",
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"group_id"},
				Description:   "The name of the user group to retrieve",
			},
			// Output fields
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the user group",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of the user group",
			},
			"attributes": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Custom attributes for the user group",
			},
			"membership_method": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Method for determining group membership (STATIC, DYNAMIC_REVIEW_REQUIRED, or DYNAMIC_AUTOMATED)",
			},
			"member_query": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Query for determining dynamic group membership",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of query",
						},
						"filter": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Filters for the query",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"field": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Field to filter on",
									},
									"operator": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Operator for the filter",
									},
									"value": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Value for the filter",
									},
								},
							},
						},
					},
				},
			},
			"member_query_exemptions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Users exempted from the dynamic group query",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the exempted user",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the exemption",
						},
					},
				},
			},
			"member_suggestions_notify": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether to send email notifications for membership suggestions",
			},
		},
	}
}

func dataSourceUserGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	var path string
	var searchType string

	// Determine search method based on provided parameters
	if groupID, ok := d.GetOk("group_id"); ok {
		// Direct lookup by ID using direct API path
		path = fmt.Sprintf("/api/v2/usergroups/%s", groupID.(string))
		searchType = "ID"
	} else if _, ok := d.GetOk("name"); ok {
		// For name, we'll get all groups and filter client-side
		path = "/api/v2/usergroups"
		searchType = "name"
	} else {
		return diag.FromErr(fmt.Errorf("one of group_id or name must be provided"))
	}

	tflog.Debug(ctx, fmt.Sprintf("Reading JumpCloud user group by %s", searchType))

	// Make API request
	resp, err := c.DoRequest(http.MethodGet, path, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading user group: %v", err))
	}

	// Handle search results vs direct ID lookup
	var group common.UserGroup
	if searchType == "ID" {
		// Direct lookup by ID returns a single group object
		if err := json.Unmarshal(resp, &group); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing user group response: %v", err))
		}
	} else {
		// For name, we get all groups and filter client-side
		// The API returns an array of groups
		var groups []common.UserGroup

		if err := json.Unmarshal(resp, &groups); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing user groups response: %v", err))
		}

		// Filter groups based on name
		var matchedGroups []common.UserGroup
		searchValue := d.Get("name").(string)

		for _, g := range groups {
			if g.Name == searchValue {
				matchedGroups = append(matchedGroups, g)
			}
		}

		if len(matchedGroups) == 0 {
			return diag.FromErr(fmt.Errorf("no user group found with name: %s", searchValue))
		}

		if len(matchedGroups) > 1 {
			tflog.Warn(ctx, fmt.Sprintf("Multiple user groups found with name: %s, using the first one", searchValue))
		}

		group = matchedGroups[0]
	}

	// Set the ID
	d.SetId(group.ID)

	// Set all the computed fields
	if err := d.Set("name", group.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %v", err))
	}

	if err := d.Set("description", group.Description); err != nil {
		return diag.FromErr(fmt.Errorf("error setting description: %v", err))
	}

	if err := d.Set("type", group.Type); err != nil {
		return diag.FromErr(fmt.Errorf("error setting type: %v", err))
	}

	// Set dynamic group fields
	if err := d.Set("membership_method", group.MembershipMethod); err != nil {
		return diag.FromErr(fmt.Errorf("error setting membership_method: %v", err))
	}

	if err := d.Set("member_suggestions_notify", group.MemberSuggestionsNotify); err != nil {
		return diag.FromErr(fmt.Errorf("error setting member_suggestions_notify: %v", err))
	}

	// Set member query if present
	if group.MemberQuery != nil {
		if err := d.Set("member_query", flattenMemberQuery(group.MemberQuery)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting member_query: %v", err))
		}
	}

	// Set member query exemptions if present
	if len(group.MemberQueryExemptions) > 0 {
		if err := d.Set("member_query_exemptions", flattenMemberQueryExemptions(group.MemberQueryExemptions)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting member_query_exemptions: %v", err))
		}
	}

	// Process attributes
	if group.Attributes != nil {
		attributes := make(map[string]interface{})
		for k, v := range group.Attributes {
			attributes[k] = fmt.Sprintf("%v", v)
		}
		if err := d.Set("attributes", attributes); err != nil {
			return diag.FromErr(fmt.Errorf("error setting attributes: %v", err))
		}
	}

	return diags
}
