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
		// Direct lookup by ID
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
		// The API returns a response with a "results" array
		var response struct {
			Results    []common.UserGroup `json:"results"`
			TotalCount int                `json:"totalCount"`
		}

		if err := json.Unmarshal(resp, &response); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing user groups response: %v", err))
		}

		// Filter groups based on name
		var matchedGroups []common.UserGroup
		searchValue := d.Get("name").(string)

		for _, g := range response.Results {
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
	d.Set("name", group.Name)
	d.Set("description", group.Description)
	d.Set("type", group.Type)

	// Process attributes
	if group.Attributes != nil {
		attributes := make(map[string]interface{})
		for k, v := range group.Attributes {
			attributes[k] = fmt.Sprintf("%v", v)
		}
		d.Set("attributes", attributes)
	}

	return diags
}
